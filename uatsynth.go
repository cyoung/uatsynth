package uatsynth

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dgryski/go-bitstream"
	"strings"
)

type UATFrame struct {
	Raw_data     []byte
	FISB_data    []byte
	FISB_month   uint32
	FISB_day     uint32
	FISB_hours   uint32
	FISB_minutes uint32
	FISB_seconds uint32

	FISB_length uint32

	frame_length uint32
	Frame_type   uint32

	Product_id uint32
	// Text data, if applicable.
	Text_data []string

	// Flags.
	a_f bool
	g_f bool
	p_f bool
	s_f bool //TODO: Segmentation.

	// For AIRMET/NOTAM.
	//FIXME: Temporary.
	//	Points             []GeoPoint
	ReportNumber       uint16
	ReportYear         uint16
	LocationIdentifier string
	RecordFormat       uint8
	ReportStart        string
	ReportEnd          string

	// For NEXRAD.
	//	NEXRAD []NEXRADBlock
}

type UATMsg struct {
	// Metadata from demodulation.
	RS_Err         int
	SignalStrength int
	// Raw message.
	msg     []byte
	Decoded bool //FIXME: There should be an indicator to show whether or not the frame has been decoded or encoded.

	// Station metadata.
	// Station location for uplink frames, aircraft position for downlink frames.
	Lat        float64
	Lon        float64
	UTCCoupled bool
	Frames     []*UATFrame
}

const (
	dlac_alpha = "\x03ABCDEFGHIJKLMNOPQRSTUVWXYZ\x1A\t\x1E\n| !\"#$%&'()*+,-./0123456789:;<=>?"
)

func dlac_encode(s string) ([]byte, error) {
	if len(s) == 0 {
		return []byte{}, nil
	}
	var buf bytes.Buffer
	bitBuf := bitstream.NewWriter(&buf)
	for i := 0; i < len(s); i++ {
		chr_num := strings.Index(dlac_alpha, string(s[i]))
		if chr_num < 0 {
			return buf.Bytes(), errors.New("can't find characters in string.")
		}
		for k := 5; k >= 0; k-- {
			bitBuf.WriteBit((chr_num>>uint(k))&0x01 != 0)
		}
	}
	// Pad with zeroes at the end.
	bitBuf.Flush(false)
	return buf.Bytes(), nil
}

//http://www.icao.int/safety/acp/inactive%20working%20groups%20library/acp-wg-c-uat-4/uat-swg04-wp05%20-%20draft%20tech%20manual-v0-4%20.pdf
func (u *UATMsg) EncodeUplink() ([][]byte, error) {
	// Allocate 8 bytes for the header.
	headerBuf := make([]byte, 8)

	// Tower location data.
	if u.Lat < 0 {
		u.Lat += 180.0
	}
	if u.Lon < 0 {
		u.Lon += 360.0
	}

	raw_lat := uint32(u.Lat * (16777216.0 / 360.0))
	raw_lon := uint32(u.Lon * (16777216.0 / 360.0))

	// Lat.
	headerBuf[0] = byte(raw_lat >> 15)
	headerBuf[1] = byte(raw_lat >> 7)
	headerBuf[2] = byte((raw_lat & 0x7F) << 1)

	// Lon.
	headerBuf[2] = headerBuf[2] | byte(raw_lon>>23)
	headerBuf[3] = byte(raw_lon >> 15)
	headerBuf[4] = byte(raw_lon >> 7)
	headerBuf[5] = byte((raw_lon & 0x7F) << 1)

	// UTC coupled, 3.2.2.1.4.
	if u.UTCCoupled {
		headerBuf[6] = headerBuf[6] | 0x80
	}

	// Application Data Valid, 3.2.2.1.6.
	//FIXME: Always true.
	headerBuf[6] = headerBuf[6] | 0x20

	// Slot ID, 3.2.2.1.7.
	//FIXME: Static slot ID of "1".
	headerBuf[6] = headerBuf[6] | 0x01

	// TIS-B Site ID, 3.2.2.1.8.
	//FIXME: Static site ID of "1".
	headerBuf[7] = 0x10

	// This buffer contains the frames (with header).
	frameBuffer := make([][]byte, 0)

	// Now begin adding in the info frames.
	for i := 0; i < len(u.Frames); i++ {
		//TODO: Implement other frame types.
		frm := u.Frames[i]
		if frm.Frame_type != 0 {
			continue // Skip all frames other than TIS-B.
		}
		switch frm.Product_id {
		case 413:
			if len(frm.Text_data) == 0 { // Nothing to do.
				continue
			}
			header := make([]byte, 2)      // Frame header.
			frm.Raw_data = make([]byte, 4) // Actual info frame, starting with info frame header.

			// Type.
			header[1] = header[1] | byte(frm.Frame_type)
			// Product ID.
			frm.Raw_data[0] = byte(frm.Product_id >> 6)
			frm.Raw_data[1] = byte((frm.Product_id & 0x1F) << 2)

			t_opt := 0
			frm.Raw_data[1] = frm.Raw_data[1] | byte(t_opt>>1)
			frm.Raw_data[2] = byte((t_opt & 0x01) << 7)
			// Hours.
			frm.Raw_data[2] = frm.Raw_data[2] | byte((frm.FISB_hours&0x1F)<<2)
			// Minutes.
			frm.Raw_data[2] = frm.Raw_data[2] | byte(frm.FISB_minutes>>4)
			frm.Raw_data[3] = byte((frm.FISB_minutes & 0x0F) << 4)

			// Encode the text data and append it.
			//FIXME. This will only take the first text report.
			encoded_data, err := dlac_encode(frm.Text_data[0])
			if err != nil {
				fmt.Printf("error encoding data: %s\n", err.Error())
				continue
			}
			frm.Raw_data = append(frm.Raw_data, encoded_data...)

			// Finalize the header length.
			// Length.
			frm.frame_length = uint32(len(frm.Raw_data))

			header[0] = byte(frm.frame_length >> 1)
			header[1] = byte((frm.frame_length & 0x01) << 7)

			// Save the data in the framebuffer.
			frameData := append(header, frm.Raw_data...)
			frameBuffer = append(frameBuffer, frameData)
		default:
			break
		}
	}

	if len(frameBuffer) == 0 { // Nothing to do.
		return nil, errors.New("nothing to encode.")
	}

	// Copy the same header multiple times, adding frames until we reach 432 bytes.
	messages := make([][]byte, 1)
	messages[0] = headerBuf
	// Generate messages up to 432 bytes long.
	for i := 0; i < len(frameBuffer); i++ {
		curMessage := len(messages) - 1
		if len(messages[curMessage])+len(frameBuffer[i]) >= 432 {
			newMessage := append(headerBuf, frameBuffer[i]...)
			messages = append(messages, newMessage)
		} else {
			messages[curMessage] = append(messages[curMessage], frameBuffer[i]...)
		}
	}

	// Pad all messages to 432 bytes.
	for i := 0; i < len(messages); i++ {
		messages[i] = append(messages[i], make([]byte, 432-len(messages[i]))...)
	}

	return messages, nil
}
