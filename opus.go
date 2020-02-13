package main

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"
	
	"layeh.com/gopus"
)

// Opus contains all the data required to keep an Opus process running and usable.
type Opus struct {
	sync.Mutex

	decoder *gopus.Decoder //The gopus decoder
	pcm16   []int16        //A buffer of 16-bit PCM samples

	closed bool
}

// NewOpus returns an initialized *Opus or an error if one could not be created.
//
// The returned *Opus can be used as both an io.Reader and an io.Writer.
// Writing Opus audio samples to the *Opus will buffer 16-bit PCM audio samples, which can then be read from the *Opus.
func NewOpus(sampleRate, channels int) (*Opus, error) {
	opusdec, err := gopus.NewDecoder(sampleRate, channels)
	if err != nil {
		return nil, err
	}

	return &Opus{
		decoder: opusdec,
		pcm16:   make([]int16, 0),
	}, nil
}

// IsRunning returns whether or not the Opus process is running, per our knowledge.
func (dec *Opus) IsRunning() bool {
	return !dec.closed
}

// Close closes the Opus session gracefully and renders the struct unusable.
func (dec *Opus) Close() {
	if dec.closed {
		return
	}

	dec.Lock()
	defer dec.Unlock()

	dec.decoder.ResetState()
	dec.decoder = nil
	dec.pcm16 = nil
	dec.closed = true
}

// Read implements an io.Reader wrapper around *Opus and reads out buffered 16-bit PCM frames in bytes.
func (dec *Opus) Read(pcm16 []byte) (n int, err error) {
	if !dec.IsRunning() {
		return 0, errors.New("opus: not running")
	}
	
	Debug.Println("--", "Length of read buffer:", len(pcm16))

	dec.Lock()
	defer dec.Unlock()

	for i := 0; i < len(pcm16); i += 2 {
		//Debug.Println("--", "Reading next PCM sample...")
		nxt, err := dec.readNextPCM()
		if err != nil {
			Error.Println(err)
			return n, err
		}
		pcm16[i] = nxt[0]
		pcm16[i+1] = nxt[1]
		n += 2
	}

	return
}

// readNext reads the next 16-bit PCM sample from the buffer, puts it in a 2-byte slice, and returns it, removing it from the buffer.
func (dec *Opus) readNextPCM() (nxt []byte, err error) {
	if len(dec.pcm16) == 0 {
		return nil, io.EOF
	}
	
	tmp16 := dec.pcm16[0]
	nxt = make([]byte, 2)
	binary.LittleEndian.PutUint16(nxt, uint16(tmp16))
	
	if len(dec.pcm16) >= 2 {
		dec.pcm16 = dec.pcm16[1:]
	} else {
		dec.pcm16 = make([]int16, 0)
	}

	return
}

// Write implements an io.Writer wrapper around *Opus and automatically decodes Opus audio samples to 16-bit PCM.
func (dec *Opus) Write(opus []byte) error {
	if !dec.IsRunning() {
		return errors.New("opus: not running")
	}
	
	Debug.Println("--", "Writing Opus sample (", len(opus), "):", opus)

	dec.Lock()
	defer dec.Unlock()

	//Decode the Opus samples to 16-bit PCM with a frame size of 960 and FEC (Forward Error Correction) disabled
	pcm16, err := dec.decoder.Decode(opus, 960, false)
	if err != nil {
		Error.Println(err)
		return err
	}
		
	//Reset the decoder state
	dec.decoder.ResetState()
		
	//Append the new PCM samples to the main buffer
	dec.pcm16 = append(dec.pcm16, pcm16...)

	return nil
}

