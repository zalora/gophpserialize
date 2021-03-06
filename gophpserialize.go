package gophpserialize

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Serializer struct {
	raw []byte
	pos int
}

func (s *Serializer) SetRaw(msg []byte) {
	s.raw = msg
}

func (s *Serializer) read() map[string]interface{} {
	r := s.readValue().(map[string]interface{})
	return r
}

func (s *Serializer) readType() byte {
	result := s.raw[s.pos]
	s.move()
	return result
}

func (s *Serializer) readBool() bool {
	result := s.raw[s.pos]
	s.move()
	if result == '0' {
		return false
	}
	return true
}

func (s *Serializer) readInt() int {
	start := s.pos
	end := start + 1

	c := s.raw[end]
	for c != ':' && c != ';' {
		end = end + 1
		c = s.raw[end]
	}
	s.pos = end
	i, _ := strconv.ParseInt(string(s.raw[start:end]), 10, 32)
	return int(i)
}

func (s *Serializer) readFloat() float64 {
	start := s.pos
	end := start + 1

	c := s.raw[end]
	for c != ':' && c != ';' {
		end = end + 1
		c = s.raw[end]
	}
	s.pos = end
	d, _ := strconv.ParseFloat(string(s.raw[start:end]), 64)
	return d
}

func (s *Serializer) readString(size int) string {
	s.move()
	result := string(s.raw[s.pos : s.pos+size])
	s.pos += size + 1
	return strings.Replace(result, "\000", "", -1)
}

func (s *Serializer) readValue() interface{} {
	objType := s.readType()
	if objType == 'N' {
		s.move()
		return nil
	}

	if objType == 'i' {
		s.move()
		val := s.readInt()
		s.move()
		return val
	}
	if objType == 'd' {
		s.move()
		val := s.readFloat()
		s.move()
		return val
	}

	if objType == 'b' {
		s.move()
		val := s.readBool()
		s.move()
		return val
	}

	if objType == 's' {
		s.move()
		size := s.readInt()
		s.move()
		val := s.readString(size)
		s.move()
		return val
	}

	if objType == 'a' {
		s.move()
		size := s.readInt()
		s.move()

		// array open {
		s.move()

		r := make(map[string]interface{})
		l := make([]interface{}, 0)

		//hack to handle array that has both string/int as key
		//convert int key to string key
		hasStringKey := false
		notAList := false

		for i := 0; i < size; i++ {
			key := s.readValue()
			val := s.readValue()
			switch v2 := key.(type) {
			case string:
				hasStringKey = true
				r[v2] = val
			case int:
				if hasStringKey || v2 != i || notAList {
					r[strconv.Itoa(v2)] = val
					if v2 != i {
						notAList = true
					}
				} else {
					l = append(l, val)
				}
			}
		}

		if len(r) > 0 && len(l) > 0 {
			for i, val := range l {
				r[strconv.Itoa(i)] = val
			}
		}

		// array close }
		s.move()
		if len(r) == 0 {
			return l
		}
		return r
	}

	if objType == 'O' {
		s.move()
		size := s.readInt()
		s.move()
		s.readString(size)
		s.move()
		objectSize := s.readInt()
		s.move()

		// object content open
		s.move()

		r := make(map[string]interface{})

		//hack to handle array that has both string/int as key

		for i := 0; i < objectSize; i++ {
			propertyName := s.readValue().(string)
			val := s.readValue()
			r[propertyName] = val
		}

		// object content close }
		s.move()
		return r
	}
	// Serialised object
	if objType == 'C' {
		s.move()
		size := s.readInt()
		s.move()
		key := s.readString(size)
		s.move()
		s.readInt()
		s.move()

		// object content open
		s.move()

		r := make(map[string]interface{})
		v := s.readValue()

		if vmap, ok := v.(map[string]interface{}); ok {
			r[key] = vmap
		}

		// object content close }
		s.move()
		return r
	}
	if objType == '{' {
		s.move()
		return nil
	}
	if objType == '}' {
		s.move()
		return nil
	}
	panic("Unknown objType: " + string(objType) + "\n" + string(s.raw))
}

func (s *Serializer) move() {
	s.pos += 1
}

func Unmarshal(data []byte) (rv interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Cant Unmarshal data %v, error: %v", data, r)
		}
	}()
	s := new(Serializer)
	s.SetRaw(data)
	return s.readValue(), nil
}

func PhpToJson(phpData []byte) (jsonData []byte, err error) {
	r, err := Unmarshal(phpData)
	if err != nil {
		return nil, err
	}
	jsonData, err = json.Marshal(r)
	return
}
