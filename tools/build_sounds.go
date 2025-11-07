package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// makeLaserSound generates a laser sound effect WAV file
func makeLaserSound(outputPath string, duration float64) error {
	const sampleRate = 44100

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	numSamples := int(sampleRate * duration)

	// Write WAV header
	if err := writeWavHeader(f, numSamples, sampleRate); err != nil {
		return fmt.Errorf("failed to write WAV header: %w", err)
	}

	// Generate laser sound effect
	// Frequency sweeps down quickly from 800Hz to 100Hz
	for i := range numSamples {
		t := float64(i) / sampleRate

		// Frequency sweep (down)
		freq := 800.0 - 700.0*t*4

		// Square wave for retro laser effect
		sample := math.Sin(2 * math.Pi * freq * t)
		if sample > 0 {
			sample = 1
		} else {
			sample = -1
		}

		// Fade out toward the end
		amp := 0.2 * (1 - t/duration)
		value := int16(sample * amp * 32767)

		// Write stereo samples (left and right channels)
		if err := binary.Write(f, binary.LittleEndian, value); err != nil {
			return fmt.Errorf("failed to write audio sample: %w", err)
		}
		if err := binary.Write(f, binary.LittleEndian, value); err != nil {
			return fmt.Errorf("failed to write audio sample: %w", err)
		}
	}

	fmt.Printf("Generated laser sound: %s (%.2fs, %d samples)\n", outputPath, duration, numSamples)
	return nil
}

// writeWavHeader writes a standard WAV file header
// I've written this code twice in my life, and believe me: it's easier in C++
func writeWavHeader(f *os.File, numSamples int, sampleRate int) error {
	const (
		numChannels   = 2 // Stereo
		bitsPerSample = 16
	)

	byteRate := sampleRate * numChannels * (bitsPerSample / 8)
	blockAlign := numChannels * (bitsPerSample / 8)
	dataSize := numSamples * blockAlign
	chunkSize := 36 + dataSize

	// RIFF header
	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(chunkSize)); err != nil {
		return err
	}
	if _, err := f.Write([]byte("WAVE")); err != nil {
		return err
	}

	// fmt chunk
	if _, err := f.Write([]byte("fmt ")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil { // PCM header length
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil { // PCM format
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(numChannels)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(byteRate)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(blockAlign)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(bitsPerSample)); err != nil {
		return err
	}

	// data chunk
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(dataSize)); err != nil {
		return err
	}

	return nil
}
