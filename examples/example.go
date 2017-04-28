package main

import (
	"fmt"
	"github.com/cyoung/ADDS"
	"github.com/cyoung/uatsynth"
)

func main() {
	var msg UATMsg

	msg.decoded = true //FIXME: There should be an indicator to show whether or not the frame has been decoded or encoded.
	msg.Lat = 42.984923
	msg.Lon = -81.245277
	msg.UTCCoupled = true //FIXME: GPS lock.

	updateList := []string{"CYCK", "CYXU", "CYKF", "CYHM", "CYYZ", "CYZR", "CWGD", "CWLS", "CYBN", "CXET", "CYVV"}

	// Get the latest METAR for each station in the list.
	for _, station := range updateList {
		metar, err := ADDS.GetLatestADDSMETAR(station)
		if err != nil {
			fmt.Printf("couldn't get METAR for '%s': %s\n", station, err.Error())
			continue
		}
		f := new(UATFrame)
		f.Text_data = append(f.Text_data, metar.Text)
		f.FISB_hours = uint32(metar.Observation.Time.Hour())
		f.FISB_minutes = uint32(metar.Observation.Time.Minute())
		f.Product_id = 413
		f.Frame_type = 0
		msg.Frames = append(msg.Frames, f)
		fmt.Printf("%02d:%02d:%s: %s\n", metar.Observation.Time.Hour(), metar.Observation.Time.Minute(), metar.Observation.Time, metar.Text)
	}

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
