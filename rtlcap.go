package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/bemasher/rtltcp"
)

const (
	BlockSize = 1 << 12
)

type Size int64

func (s Size) String() string {
	return strconv.FormatInt(int64(s), 10)
}

func (s *Size) Set(value string) (err error) {
	var (
		mantissa float64
		exponent string
	)

	mantissa, err = strconv.ParseFloat(value, 64)

	if err == nil {
		*s = Size(mantissa)
		return
	}

	_, err = fmt.Sscanf(value, "%f%s", &mantissa, &exponent)
	if err != nil {
		return
	}

	switch strings.ToLower(exponent) {
	case "k":
		*s = Size(mantissa * (1 << 10))
	case "m":
		*s = Size(mantissa * (1 << 20))
	case "g":
		*s = Size(mantissa * (1 << 30))
	case "t":
		*s = Size(mantissa * (1 << 40))
	case "p":
		*s = Size(mantissa * (1 << 50))
	case "e":
		*s = Size(mantissa * (1 << 60))
	case "z":
		*s = Size(mantissa * (1 << 70))
	case "y":
		*s = Size(mantissa * (1 << 80))
	default:
		err = fmt.Errorf("invalid expontent")
	}

	return
}

var (
	size      Size
	timeLimit time.Duration
	squelch   float64
	filename  string
)

// Default Magnitude Lookup Table
type MagLUT []float64

// Pre-computes normalized squares with most common DC offset for rtl-sdr dongles.
func NewSqrtMagLUT() (lut MagLUT) {
	lut = make([]float64, 0x100)
	for idx := range lut {
		lut[idx] = (127.4 - float64(idx)) / 127.6
		lut[idx] *= lut[idx]
	}
	return
}

// Calculates complex magnitude on given IQ stream writing result to output.
func (lut MagLUT) Execute(input []byte, output []float64) {
	for idx := 0; idx < len(input); idx += 2 {
		output[idx>>1] = math.Sqrt(lut[input[idx]] + lut[input[idx+1]])
	}
}

func Mean(sig []float64) (mean float64) {
	for _, val := range sig {
		mean += val
	}
	mean /= float64(len(sig))
	return
}

func init() {
	log.SetFlags(log.Lmicroseconds)
	rand.Seed(time.Now().UnixNano())

	flag.Var(&size, "bytes", "number of bytes to capture")
	flag.DurationVar(&timeLimit, "duration", 0, "length of time to capture")
	flag.Float64Var(&squelch, "squelch", 0.0, "minimum mean level a sample block must be to commit to disk")
	flag.StringVar(&filename, "o", "samples.bin", "filename to write samples to")
}

func main() {
	var sdr rtltcp.SDR
	sdr.RegisterFlags()

	flag.Parse()

	err := sdr.Connect(nil)
	if err != nil {
		log.Fatal(err)
	}

	defer sdr.Close()

	sdr.HandleFlags()

	sampleFile, err := os.Create(filename)
	if err != nil {
		log.Fatal("error creating output file:", err)
	}
	defer sampleFile.Close()

	var (
		tLimit    <-chan time.Time
		sigint    chan os.Signal
		bytesRead int64
	)
	if timeLimit != 0 {
		sigint = make(chan os.Signal, 1)
		signal.Notify(sigint, os.Kill, os.Interrupt)
		tLimit = time.After(timeLimit)
	}

	block := make([]byte, BlockSize)
	mag := make([]float64, BlockSize>>1)

	lut := NewSqrtMagLUT()

	var min, max float64
	meanTick := time.Tick(time.Second)

	for {
		// Exit on interrupt or time limit, otherwise receive.
		select {
		case <-sigint:
			return
		case <-tLimit:
			return
		case <-meanTick:
			log.Printf("Min: %0.3f Max: %0.3f\n", min, max)
			min = math.MaxFloat64
			max = -math.MaxFloat64
		default:
			if size != 0 && bytesRead > int64(size) {
				return
			}

			n, err := sdr.Read(block)
			if err != nil {
				log.Fatal("Error reading sample block:", err)
			}

			lut.Execute(block, mag)
			mean := Mean(mag)
			if mean > max {
				max = mean
			}
			if mean < min {
				min = mean
			}

			if squelch != 0 && mean < squelch {
				continue
			}

			bytesRead += int64(n)

			_, err = sampleFile.Write(block)
			if err != nil {
				log.Fatal("Error writing sample block:", err)
			}
		}
	}
}
