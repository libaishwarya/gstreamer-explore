package main

import (
	"fmt"
	"log"

	"github.com/go-gst/go-gst/gst"
)

func main() {
	// Initialize GStreamer
	gst.Init(nil)
	fmt.Println("✅ GStreamer initialized")

	// Create a new pipeline
	pipeline, err := gst.NewPipeline("audio-mixer")
	if err != nil {
		log.Fatalf("Failed to create pipeline: %v", err)
	}
	fmt.Println("✅ Pipeline created")

	// Create elements
	src1, err := gst.NewElement("audiotestsrc")
	if err != nil {
		log.Fatalf("Failed to create audiotestsrc: %v", err)
	}
	src1.SetProperty("wave", 4) // Sine wave
	fmt.Println("✅ audiotestsrc (sine wave) created")

	src2, err := gst.NewElement("autoaudiosrc")
	if err != nil {
		log.Fatalf("Failed to create autoaudiosrc: %v", err)
	}
	fmt.Println("✅ autoaudiosrc (microphone) created")

	convert1, err := gst.NewElement("audioconvert")
	if err != nil {
		log.Fatalf("Failed to create audioconvert: %v", err)
	}

	convert2, err := gst.NewElement("audioconvert")
	if err != nil {
		log.Fatalf("Failed to create audioconvert: %v", err)
	}

	level, err := gst.NewElement("level")
	if err != nil {
		log.Fatalf("Failed to create level: %v", err)
	}
	level.SetProperty("interval", uint64(50000000)) // 50ms interval
	level.SetProperty("message", true)              // Enable level messages
	fmt.Println("✅ level element created (for voice detection)")

	volume1, err := gst.NewElement("volume") // Controls sine wave
	if err != nil {
		log.Fatalf("Failed to create volume1: %v", err)
	}
	volume1.SetProperty("volume", 1.0) // Default volume

	mixer, err := gst.NewElement("audiomixer")
	if err != nil {
		log.Fatalf("Failed to create audiomixer: %v", err)
	}

	sink, err := gst.NewElement("autoaudiosink")
	if err != nil {
		log.Fatalf("Failed to create autoaudiosink: %v", err)
	}

	// Add elements to the pipeline
	pipeline.AddMany(src1, convert1, level, volume1, src2, convert2, mixer, sink)

	// Link elements
	if err := gst.ElementLinkMany(src1, convert1, level, volume1, mixer); err != nil {
		log.Fatalf("Failed to link elements for src1: %v", err)
	}
	if err := gst.ElementLinkMany(src2, convert2, mixer); err != nil {
		log.Fatalf("Failed to link elements for src2: %v", err)
	}
	if err := gst.ElementLinkMany(mixer, sink); err != nil {
		log.Fatalf("Failed to link mixer to sink: %v", err)
	}

	fmt.Println("✅ All elements linked successfully")

	// Set up a bus to listen for messages
	bus := pipeline.GetPipelineBus()
	// TODO: Find why the message is not getting received for detecting voice.
	bus.AddWatch(func(msg *gst.Message) bool {
		log.Printf("Received Message: %v", msg.Type()) // Log all message types
		switch msg.Type() {
		case gst.MessageError:
			err := msg.ParseError()
			log.Printf("❌ ERROR: %v", err)
		case gst.MessageElement:
			s := msg.GetStructure()
			if s != nil {
				log.Printf("Element Message Structure: %v", s.Name())
				if s.Name() == "level" {
					v, err := s.GetValue("rms")
					if err != nil {
						log.Printf("Failed to get RMS value: %v", err)
						return true
					}
					rms, ok := v.([]float64)
					if !ok || len(rms) == 0 {
						log.Printf("Invalid RMS value: %v", v)
						return true
					}
					fmt.Printf("\n🔊 RMS Level: %.2f dB", rms[0])

					if rms[0] > -30.0 {
						fmt.Println(" - Speech detected, muting sine wave")
						volume1.SetProperty("volume", 0.0) // Mute sine wave
					} else {
						fmt.Println(" - No speech detected, playing sine wave")
						volume1.SetProperty("volume", 1.0) // Unmute sine wave
					}
				}
			}
		case gst.MessageStateChanged:
			oldState, newState := msg.ParseStateChanged()
			log.Printf("Pipeline state changed from %s to %s", oldState, newState)
			if newState == gst.StatePlaying {
				fmt.Println("🚀 Pipeline is now playing!")
			}
		}
		return true
	})

	// Start the pipeline
	err = pipeline.SetState(gst.StatePlaying)
	if err != nil {
		log.Fatalf("Failed to start pipeline: %v", err)
	}
	fmt.Println("🚀 Pipeline started, listening for audio...")

	// Block the main thread to keep the application running
	select {}
}
