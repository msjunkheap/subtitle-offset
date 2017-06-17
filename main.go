package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

var (
	args   []string
	offset time.Duration
)

func init() {
	flag.DurationVar(&offset, "o", 0, "Offset to apply")
	flag.Parse()
	args = flag.Args()

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Please provide subtitle filename\n")
		os.Exit(1)
	}
}

func main() {
	file, err := ioutil.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read subtitle file: %s %v", args[0], err)
		os.Exit(1)
	}

	if time.Duration(0) == offset {
		fmt.Fprintf(os.Stderr, "Offset is zero, file unmodified\n")
		fmt.Fprintf(os.Stdout, "%s", file)
		os.Exit(0)
	}

	r := regexp.MustCompile(`(\d{2}:\d{2}:\d{2},\d{3})\s-->\s(\d{2}:\d{2}:\d{2},\d{3})`)
	indexes := r.FindAllSubmatchIndex(file, -1)
	if len(indexes) == 0 {
		fmt.Fprintf(os.Stderr, "Found no subtitle timestamps, unable to adjust offset\n")
		os.Exit(2)
	}

	// Instead of allocating more memory, we'll just overwrite the old data with a slice backed
	// by the original file array
	newFile := file[:0]

	var last int
	for _, i := range indexes {
		// Fill with the original data up until this point
		newFile = append(newFile, file[last:i[2]]...)

		// Overwrite the old subtitle start timestamp with the new one
		newFile = append(newFile, adjust(offset, file[i[2]:i[3]])...)

		// Write back the filler content between subtitle start/end
		newFile = append(newFile, file[i[3]:i[4]]...)

		// Overwrite the old subtitle end timestamp with the new one
		newFile = append(newFile, adjust(offset, file[i[4]:i[5]])...)

		last = i[5]
	}

	fmt.Fprintf(os.Stdout, "%s", file)
}

func adjust(o time.Duration, match []byte) []byte {
	// Switch the comma to a period so we can parse it
	match[8] = '.'
	// Parse the timestamp
	t, err := time.Parse("15:04:05.999", string(match[:]))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Something went wrong: %v\n", err)
		os.Exit(3)
	}

	// Apply the offset time
	t = t.Add(o)

	h := t.Hour()
	m := t.Minute()
	s := t.Second()
	ms := (t.Nanosecond() / 1000000) % 1000

	return []byte(fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms))
}
