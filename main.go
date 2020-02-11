package main

/*
#cgo LDFLAGS: ${SRCDIR}/libspeexdsp.a -lm
#include "speex/speex_echo.h"
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>

#define NN 128
#define TAIL 1024

void print(short *c, int number){
	for (int i = 0; i < number; i++){
		printf("%d\n", *(c+i));
	}
}

short * AEC(short *mic, short *speaker, int signalSize, int sampleRate){
	printf("AEC processing...\n");

	short *res = malloc(signalSize);
	SpeexEchoState *st = speex_echo_state_init(NN, TAIL);
	speex_echo_ctl(st, SPEEX_ECHO_SET_SAMPLING_RATE, &sampleRate);

	//echo_cancel
	for (int i = 0; i < signalSize / 2; i+= NN) {
		short *frameMic = malloc(NN * sizeof(short));
		short *frameSpeaker = malloc(NN * sizeof(short));
		short *frameOutput = malloc(NN * sizeof(short));

		for (int j = 0; j < NN; j++){
			frameMic[j] = mic[i + j];
			frameSpeaker[j] = speaker[i + j];
		}

		speex_echo_cancellation(st, frameMic, frameSpeaker, frameOutput);

		//Write result to file .raw
		//fwrite(frameOutput, sizeof(short), NN, out);

		for (int j = 0; j < NN; j++) {
			res[j + i] = frameOutput[j];
		}
		free(frameMic);
		free(frameSpeaker);
		free(frameOutput);
	}
	//end

	speex_echo_state_destroy(st);
	printf("AEC done!\n");
	return res;
}

*/
import "C"
import (
	"fmt"
	"github.com/timshannon/go-openal/openal"
	"io/ioutil"
	"time"
	"unsafe"
)

const SAMPLE_RATE = 44100

func byteToInt16(input []byte) []int16 {
	res := make([]int16, len(input)/2)
	for i := 0; i < len(input); i += 2 {
		res[i/2] = int16(int(input[i+1])<<8 + int(input[i])&0xFF)
	}
	return res
}

func int16ToByte(input []int16) []byte {
	res := make([]byte, len(input)*2)
	for i := 0; i < len(input); i++ {
		res[i*2] = byte(input[i])
		res[i*2+1] = byte(input[i] >> 8)
	}
	return res
}

func goAEC(mic []byte, speaker []byte, signalSize int, sampleRate int) []byte {
	sMic := byteToInt16(mic)
	sSpeaker := byteToInt16(speaker)

	var resFromC *C.short = C.AEC((*C.short)(unsafe.Pointer(&sMic[0])), (*C.short)(unsafe.Pointer(&sSpeaker[0])), (C.int)(signalSize), (C.int)(sampleRate))

	int16FromCshort := (*[1 << 28]int16)(unsafe.Pointer(resFromC))[: signalSize/2 : signalSize/2]
	res := int16ToByte(int16FromCshort)
	return res
}

func play(input []byte) {
	device := openal.OpenDevice("")
	defer device.CloseDevice()
	context := device.CreateContext()
	defer context.Destroy()
	context.Activate()
	vendor := openal.GetVendor()
	// make sure things have gone well
	if err := openal.Err(); err != nil {
		fmt.Printf("Failed to setup OpenAL: %v\n", err)
		return
	}
	fmt.Printf("OpenAL vendor: %s\n", vendor)

	source := openal.NewSource()
	defer source.Pause()
	source.SetLooping(false)
	source.SetPosition(&openal.Vector{0.0, 0.0, -5.0})
	soundBuffer := openal.NewBuffer()
	if err := openal.Err(); err != nil {
		fmt.Printf("OpenAL buffer creation failed: %v\n", err)
		return
	}
	soundBuffer.SetData(openal.FormatMono16, input, SAMPLE_RATE)
	source.SetBuffer(soundBuffer)
	source.Play()
	for source.State() == openal.Playing {
		// loop long enough to let the wave file finish
		time.Sleep(time.Millisecond * 10)
	}

	source.Delete()
}

func main() {
	mic, _ := ioutil.ReadFile("/home/ngoctb3/Desktop/INPUTMICWITHNOISE.raw")
	speaker, _ := ioutil.ReadFile("/home/ngoctb3/Desktop/INPUTMIC.raw")
	res := goAEC(mic, speaker, len(mic), SAMPLE_RATE)

	play(res)
}
