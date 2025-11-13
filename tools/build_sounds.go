package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
)

// makeLaserSound generates a laser sound effect WAV file
func makeLaserSound(outputPath string, duration float64) error {
	const sampleRate = 44100
	numSamples := int(sampleRate * duration)

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

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

func makeImpactSound(outputPath string, duration float64) error {
	const sampleRate = 44100
	numSamples := int(sampleRate * duration)

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Write WAV header
	if err := writeWavHeader(f, numSamples, sampleRate); err != nil {
		return fmt.Errorf("failed to write WAV header: %w", err)
	}

	for i := range numSamples {
		t := float64(i) / sampleRate

		// Two combined frequencies: impact + resonance
		freq1 := 1200.0 - 800.0*t*3  // downward pitch for impact
		freq2 := 4000.0 - 2000.0*t*2 // metallic ring
		sample := 0.6*math.Sin(2*math.Pi*freq1*t) +
			0.4*math.Sin(2*math.Pi*freq2*t)

		// Exponential fade to give quick decay
		env := math.Exp(-6 * t / duration)
		value := int16(sample * env * 32767)

		binary.Write(f, binary.LittleEndian, value)
		binary.Write(f, binary.LittleEndian, value)
	}

	fmt.Printf("Generated impact sound: %s (%.2fs, %d samples)\n", outputPath, duration, numSamples)
	return nil
}

func makeExplosionSound(outputPath string, duration float64) error {
	const sampleRate = 44100
	numSamples := int(sampleRate * duration)

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Write WAV header
	if err := writeWavHeader(f, numSamples, sampleRate); err != nil {
		return fmt.Errorf("failed to write WAV header: %w", err)
	}

	// Simple low-pass filtered noise explosion
	var prev float64
	for i := range numSamples {
		t := float64(i) / sampleRate

		// Envelope: fast attack, exponential decay
		env := math.Exp(-4 * t / duration)
		if t < 0.02 {
			env = t / 0.02 // attack ramp
		}

		// Random white noise
		n := (rand.Float64()*2 - 1)

		// Low-pass filter (to give a "boom" instead of hiss)
		smooth := prev*0.9 + n*0.1
		prev = smooth

		// Modulate slight pitch oscillation
		mod := math.Sin(2 * math.Pi * 30 * t)

		sample := smooth * env * (0.8 + 0.2*mod)
		value := int16(sample * 32767)

		binary.Write(f, binary.LittleEndian, value)
		binary.Write(f, binary.LittleEndian, value)
	}
	fmt.Println("Generated explosion sound: explosion.wav")
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
