package mfer

import (
	"encoding/binary"
	"errors"
	"fmt"
)

func Anonymize(bytes []byte) ([]byte, error) {
	var (
		tagCode byte
		length  uint32
	)

	// fmt.Println("len(bytes) = ", len(bytes))

	for i := 0; i < len(bytes); {
		tagCode = bytes[i]
		// fmt.Println("tagCode = ", tagCode)
		i++
		if tagCode == ZERO {
			continue
		} else if tagCode == END {
			break
		}

		length = uint32(bytes[i])
		i++

		if length > 0x7f { /* MSBが1ならば */
			numBytes := length - 0x80
			if numBytes > 4 {
				fmt.Println("byets = ", bytes[i-2:i+15])
				fmt.Printf("length = %x, numBytes = %d, bytes = %d\n", length, numBytes, bytes[i-1])
				return bytes, errors.New("error nbytes")
			}
			length = binary.BigEndian.Uint32(append(make([]byte, 4-numBytes), bytes[i:i+int(numBytes)]...))
			i += int(numBytes)
		}

		switch tagCode {
		/*
		 * for Mfer.WaveFrom
		 */
		case CHANNEL_ATTRIBUTE:
			// fmt.Println("Channel Attribute")
			length = uint32(bytes[i])
			i++

		/*
		 * for Mfer.Helper
		 */
		case P_NAME: // 患者の名前
			// fmt.Println("Patiant Name")
			bytes = append(bytes[:i-2], bytes[i+int(length):]...)
			i -= 2
			continue

		case P_ID:
			// 一旦IDも削除する実装
			// ここでなんかハッシュ化したい
			// fmt.Println("Patiant ID")
			bytes = append(bytes[:i-2], bytes[i+int(length):]...)
			i -= 2
			continue

		case P_AGE:
			//fmt.Println("Patiant Age")
			// 一旦削除
			bytes = append(bytes[:i-2], bytes[i+int(length):]...)
			i -= 2
			continue
			/*
				mfer.Helper.Patient.Age = bytes[i]
				ageInDays, err := Binary2Uint32(mfer.Control.ByteOrder, bytes[i+1:i+3]...)
				birthYear, err := Binary2Uint32(mfer.Control.ByteOrder, bytes[i+3:i+5]...)
				mfer.Helper.Patient.BirthYear  = birthYear
				mfer.Helper.Patient.BirthMonth = bytes[i+5]
				mfer.Helper.Patient.BirthDay   = bytes[i+6]
			*/

			// ここの処理がうまくいかない可能性がある
			// fmt.Println("before = ", bytes[i-2:i+7])

			// この下のcopyはP_AGEのデータが7バイトで記述されている前提じゃないと動かない．
			// copy(bytes[i+1:i+3], []byte{0x00, 0x00})
			// copy(bytes[i+6:i+7], []byte{0x00}) // 生年月日から日を削除(生年月は残す)

		case P_SEX:
			// fmt.Println("Patiant Gender")
			//mfer.Helper.Patient.Sex = bytes[i]
		}
		i += int(length)
	}
	return bytes, nil
}
