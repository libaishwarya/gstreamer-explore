package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-gst/go-gst/gst"
)

func main() {
	// Initialize GStreamer
	gst.Init(nil)

	// Create a new pipeline
	pipeline, err := gst.NewPipeline("audio-mixer")
	if err != nil {
		log.Fatalf("Failed to create pipeline: %v", err)
	}

	// Create elements
	src1, err := gst.NewElement("audiotestsrc")
	if err != nil {
		log.Fatalf("Failed to create audiotestsrc: %v", err)
	}
	src1.SetProperty("wave", 4) // Sine wave

	src2, err := gst.NewElement("autoaudiosrc")
	if err != nil {
		log.Fatalf("Failed to create autoaudiosrc: %v", err)
	}

	convert1, err := gst.NewElement("audioconvert")
	if err != nil {
		log.Fatalf("Failed to create audioconvert: %v", err)
	}

	convert2, err := gst.NewElement("audioconvert")
	if err != nil {
		log.Fatalf("Failed to create audioconvert: %v", err)
	}

	vad, err := gst.NewElement("voiceactivitydetection")
	if err != nil {
		log.Fatalf("Failed to create voiceactivitydetection: %v", err)
	}

	volume, err := gst.NewElement("volume")
	if err != nil {
		log.Fatalf("Failed to create volume: %v", err)
	}

	mixer, err := gst.NewElement("audiomixer")
	if err != nil {
		log.Fatalf("Failed to create audiomixer: %v", err)
	}

	sink, err := gst.NewElement("autoaudiosink")
	if err != nil {
		log.Fatalf("Failed to create autoaudiosink: %v", err)
	}

	// Add elements to the pipeline
	pipeline.AddMany(src1, src2, convert1, convert2, vad, volume, mixer, sink)

	// Link elements
	if err := gst.ElementLinkMany(src1, convert1, vad, volume, mixer); err != nil {
		log.Fatalf("Failed to link elements for src1")
	}
	if err := gst.ElementLinkMany(src2, convert2, mixer); err != nil {
		log.Fatalf("Failed to link elements for src2")
	}
	if err := gst.ElementLinkMany(mixer, sink); err != nil {
		log.Fatalf("Failed to link mixer to sink")
	}

	// Set up a bus to listen for messages
	bus := pipeline.GetPipelineBus()
	bus.AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		// https://gstreamer.freedesktop.org/documentation/gstreamer/gstmessage.html?gi-language=c
		// 2 for error
		case 2:
			err := msg.ParseError()
			log.Printf("Error: %v", err)
		case 64:
			// 64 for state change
			// Parse the state change message
			oldState, newState := msg.ParseStateChanged()
			log.Printf("State changed: %s -> %s", oldState, newState)
			// if msg.Source() == pipeline.GetPipeline() {
			// 	oldState, newState := msg.ParseStateChanged()
			// 	log.Printf("Pipeline state changed from %s to %s", oldState, newState)
			// }
		}
		return true
	})

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	// Main loop to handle VAD events
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			activityI, err := vad.GetProperty("voice-activity")
			if err != nil {
				fmt.Printf("\nerror: %v", err)
			}
			activity := activityI.(bool)
			if activity {
				volume.SetProperty("volume", 0.0) // Mute test tone
			} else {
				volume.SetProperty("volume", 1.0) // Unmute test tone
			}
		}
	}()

	// Block the main thread to keep the application running
	select {}
}
