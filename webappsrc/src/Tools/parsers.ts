function UnescapeHelp(line: string): string {
  let result = '';
  let slash = false;

  for (let c = 0; c < line.length; ++c) {
    const char = line.charAt(c);
    if (slash) {
      if (char === '\\') {
        result += '\\';
      } else if (char === 'n') {
        result += '\n';
      } else {
        result += ('\\' + char);
      }
      slash = false;
    } else {
      if (char === '\\') {
        slash = true;
      } else {
        result += char;
      }
    }
  }

  if (slash) {
    result += '\\';
  }

  return result;
}

function BufferToFloatArray(data: ArrayBuffer): number[] {
  const out: number[] = [];

  const v = new Float32Array(data);

  for (let i = 0; i < v.length; i++) {
    out.push(v[i]);
  }

  return out;
}

function ShallowObjectEquals(obj1: any | void | null, obj2: any | void | null): boolean {
  if (!obj1 || !obj2) {
    return obj1 === obj2;
  }
  if (obj1 === obj2) {
    return true;
  }
  const obj1Keys = Object.keys(obj1);
  if (obj1Keys.length !== Object.keys(obj2).length) {
    return false;
  }
  for (let i = 0; i < obj1Keys.length; i++) {
    const key = obj1Keys[i];
    if (obj1[key] !== obj2[key]) {
      return false;
    }
  }
  return true;
}

export {
  BufferToFloatArray,
  UnescapeHelp,
  ShallowObjectEquals,
}
