package rpc

import (
	"context"
	"github.com/racerxdl/qo100-dedrift/config"
	"time"
)

func (s *Server) Login(ctx context.Context, lr *LoginRequest) (*LoginData, error) {
	if lr.GetUsername() != s.username {
		return &LoginData{
			Status: StatusType_Invalid,
		}, nil
	}

	if !ComparePassword(lr.GetSalt(), lr.GetHashedPassword(), lr.GetTimestamp(), s.password) {
		return &LoginData{
			Status: StatusType_Invalid,
		}, nil
	}

	session := SessionData{
		expiration: time.Now().Add(time.Hour),
		loginAddr:  "TODO",
	}

	token := s.addSession(session)

	return &LoginData{
		Status: StatusType_OK,
		Token:  token,
	}, nil

}

func (s *Server) SetCenterFrequency(ctx context.Context, c *FrequencyData) (*ConfigReturn, error) {
	if !s.checkSession(c.Token) {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: "Invalid Token",
		}, nil
	}
	err := s.rc.SetCenterFrequency(uint32(c.Frequency))
	if err != nil {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: err.Error(),
		}, nil
	}

	return &ConfigReturn{
		Status:  StatusType_OK,
		Message: "OK",
	}, nil
}

func (s *Server) SetGain(ctx context.Context, c *GainData) (*ConfigReturn, error) {
	if !s.checkSession(c.Token) {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: "Invalid Token",
		}, nil
	}

	err := s.rc.SetGain(c.Gain)
	if err != nil {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: err.Error(),
		}, nil
	}

	return &ConfigReturn{
		Status:  StatusType_OK,
		Message: "OK",
	}, nil
}

func (s *Server) SetBeaconOffset(ctx context.Context, c *FrequencyData) (*ConfigReturn, error) {
	if !s.checkSession(c.Token) {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: "Invalid Token",
		}, nil
	}

	err := s.rc.SetBeaconOffset(float32(c.Frequency))
	if err != nil {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: err.Error(),
		}, nil
	}

	return &ConfigReturn{
		Status:  StatusType_OK,
		Message: "OK",
	}, nil
}

func (s *Server) SetFullFFTConfig(ctx context.Context, c *FFTConfigData) (*ConfigReturn, error) {
	if !s.checkSession(c.Token) {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: "Invalid Token",
		}, nil
	}
	err := s.rc.SetFullFFTConfig(config.FFTWindowSetting{
		MaxVal: int(c.MaxVal),
		Range:  int(c.Range),
		Width:  int(c.Width),
		Height: int(c.Height),
	})
	if err != nil {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: err.Error(),
		}, nil
	}

	return &ConfigReturn{
		Status:  StatusType_OK,
		Message: "OK",
	}, nil
}

func (s *Server) SetSegFFTConfig(ctx context.Context, c *FFTConfigData) (*ConfigReturn, error) {
	if !s.checkSession(c.Token) {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: "Invalid Token",
		}, nil
	}

	err := s.rc.SetSegFFTConfig(config.FFTWindowSetting{
		MaxVal: int(c.MaxVal),
		Range:  int(c.Range),
		Width:  int(c.Width),
		Height: int(c.Height),
	})
	if err != nil {
		return &ConfigReturn{
			Status:  StatusType_Error,
			Message: err.Error(),
		}, nil
	}

	return &ConfigReturn{
		Status:  StatusType_OK,
		Message: "OK",
	}, nil
}
