package main

import (
	"fmt"
)

func main() {
	var msg UATMsg

	msg.decoded = true //FIXME: There should be an indicator to show whether or not the frame has been decoded or encoded.
	msg.Lat = 42.984923
	msg.Lon = -81.245277
	msg.UTCCoupled = true //FIXME: GPS lock.

	// Add a frame with a single METAR.
	f := new(UATFrame)
	f.Text_data = append(f.Text_data, "METAR CYCK 152100Z AUTO 33019G24KT 9SM CLR M03/M14 A3008 RMK SLP197")
	f.FISB_hours = 18
	f.FISB_minutes = 29
	f.Product_id = 413
	f.Frame_type = 0

	msg.Frames = append(msg.Frames, f)

	// Add a frame with a single METAR.
	f2 := new(UATFrame)
	f2.Text_data = append(f2.Text_data, "METAR CYXU 152100Z 31016G31KT 2SM -SHSN BLSN FEW014 BKN023TCU OVC210 M05/M10 A3000 RMK SN1CU1TCU5CI1 VIS VRB 1-3 SLP177")
	f2.FISB_hours = 18
	f2.FISB_minutes = 29
	f2.Product_id = 413
	f2.Frame_type = 0

	msg.Frames = append(msg.Frames, f2)

	// Add a frame with a single METAR.
	f3 := new(UATFrame)
	f3.Text_data = append(f3.Text_data, "METAR CYKF 152136Z AUTO 30021G31KT 9SM -SN BKN034 BKN200 M05/M12 A2992 RMK SLP151")
	f3.FISB_hours = 18
	f3.FISB_minutes = 29
	f3.Product_id = 413
	f3.Frame_type = 0

	msg.Frames = append(msg.Frames, f3)

	// Add a frame with a single METAR.
	f4 := new(UATFrame)
	f4.Text_data = append(f4.Text_data, "METAR CYHM 152100Z 30017G29KT 8SM DRSN FEW040 BKN230 M05/M13 A2990 RMK CU1CI5 CONTRAILS SLP140")
	f4.FISB_hours = 18
	f4.FISB_minutes = 29
	f4.Product_id = 413
	f4.Frame_type = 0

	msg.Frames = append(msg.Frames, f4)

	// Add a frame with a single METAR.
	f5 := new(UATFrame)
	f5.Text_data = append(f5.Text_data, "METAR CYYZ 152100Z 34022G35KT 15SM FEW050 FEW120 OVC200 M03/M15 A2987 RMK SC1AC1CS6 SLP128")
	f5.FISB_hours = 18
	f5.FISB_minutes = 29
	f5.Product_id = 413
	f5.Frame_type = 0

	msg.Frames = append(msg.Frames, f5)

	// Add a frame with a single METAR.
	f6 := new(UATFrame)
	f6.Text_data = append(f6.Text_data, "METAR Y47 152134Z AUTO 32012G17KT 10SM BKN049 M01/M12 A3013 RMK AO2 T10111120")
	f6.FISB_hours = 18
	f6.FISB_minutes = 29
	f6.Product_id = 413
	f6.Frame_type = 0

	msg.Frames = append(msg.Frames, f6)

	encodedMessages, err := msg.EncodeUplink()
	if err != nil {
		fmt.Printf("error encoding: %s\n", err.Error())
		return
	}

	for _, m := range encodedMessages {
		fmt.Printf("+")
		for i := 0; i < len(m); i++ {
			fmt.Printf("%02x", m[i])
		}
		fmt.Printf(";\n")
	}

}
