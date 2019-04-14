package rtltcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const readTimeout = time.Second * 2

type OnSamples func([]complex64)

type Client struct {
	running    bool
	conn       net.Conn
	dongleInfo DongleInfo
	cb         OnSamples
	stopChan   chan bool

	samplesBuffer    []byte
	samplesBufferPos int
}

func MakeClient() *Client {
	return &Client{
		stopChan: make(chan bool),
		running:  false,
		dongleInfo: DongleInfo{
			TunerType: RtlsdrTunerUnknown,
			Magic:     [4]uint8{0, 0, 0, 0},
		},
		samplesBufferPos: 0,
		samplesBuffer:    make([]byte, 16384),
	}
}

func (client *Client) GetDongleInfo() DongleInfo {
	return client.dongleInfo
}

func (client *Client) SetOnSamples(cb OnSamples) {
	client.cb = cb
}

func (client *Client) SetGain(gain uint8) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(gain))

	cmd := Command{
		Type:  SetGain,
		Param: [4]byte{buff[0], buff[1], buff[2], buff[3]},
	}

	return client.SendCommand(cmd)
}

func (client *Client) SetSampleRate(sampleRate uint32) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, sampleRate)

	cmd := Command{
		Type:  SetSampleRate,
		Param: [4]byte{buff[0], buff[1], buff[2], buff[3]},
	}

	return client.SendCommand(cmd)
}

func (client *Client) SetCenterFrequency(centerFrequency uint32) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, centerFrequency)

	cmd := Command{
		Type:  SetFrequency,
		Param: [4]byte{buff[0], buff[1], buff[2], buff[3]},
	}

	return client.SendCommand(cmd)
}

func (client *Client) SendCommand(cmd Command) error {
	buffer := &bytes.Buffer{}
	err := binary.Write(buffer, binary.BigEndian, &cmd)
	if err != nil {
		return err
	}

	_, err = client.conn.Write(buffer.Bytes())
	return err
}

func (client *Client) Connect(address string) error {
	log.Debug("Connecting to %s", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	client.conn = conn
	log.Debug("Waiting for handshake")
	err = client.handshake()

	if err != nil {
		_ = conn.Close()
		return err
	}

	log.Debug("Got handshake. Running")
	client.running = true
	go client.loop()

	return nil
}

func (client *Client) Stop() {
	if client.running {
		client.running = false
		_ = client.conn.Close()
	}
}

func (client *Client) handshake() error {
	buffer := make([]byte, DongleInfoSize)

	err := client.conn.SetReadDeadline(time.Now().Add(readTimeout))
	if err != nil {
		return err
	}

	n, err := client.conn.Read(buffer)
	if err != nil {
		return err
	}

	if n != len(buffer) {
		return fmt.Errorf("not received enough bytes for handshake")
	}

	b := bytes.NewReader(buffer)
	err = binary.Read(b, binary.BigEndian, &client.dongleInfo)
	if err != nil {
		return err
	}

	log.Debug("Received Handshake. Tuner Type: %s", TunerTypeToName[client.dongleInfo.TunerType])

	return nil
}

func (client *Client) loop() {
	buffer := make([]byte, 512)

	for client.running {
		client.conn.SetReadDeadline(time.Now().Add(readTimeout))
		n, err := client.conn.Read(buffer)

		if err != nil {
			log.Error("Error reading data: %s", err)
			client.running = false
			break
		}

		o := buffer[:n]
		if len(o) > 0 {
			copy(client.samplesBuffer[client.samplesBufferPos:], o)
			client.samplesBufferPos += len(o)
		}

		if client.samplesBufferPos == len(client.samplesBuffer) {
			client.handleData(client.samplesBuffer)
			client.samplesBufferPos = 0
		}
	}

	_ = client.conn.Close()
}

func (client *Client) handleData(data []byte) {
	if client.cb != nil {
		iq := make([]complex64, len(data)/2)

		for i := range iq {
			rv := (float32(data[i*2]) - 128) / 127
			iv := (float32(data[i*2+1]) - 128) / 127
			iq[i] = complex(rv, iv)
		}

		client.cb(iq)
	}
}
