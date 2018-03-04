### Purpose
To implement an rtl-sdr signal capture tool similar to `rtl_sdr` but using `rtl_tcp` instead of direct access to the dongle. We also include a squelch mechanism for ignoring blocks of samples that don't meet a specified power level.

[![AGPLv3 License](http://img.shields.io/badge/license-AGPLv3-blue.svg?style=flat)](http://choosealicense.com/licenses/agpl-3.0/)

### Requirements
 * GoLang >=1.2 (Go build environment setup guide: http://golang.org/doc/code.html)
 * rtl-sdr
   * Windows: [pre-built binaries](http://sdr.osmocom.org/trac/attachment/wiki/rtl-sdr/RelWithDebInfo.zip)
   * Linux: [source and build instructions](http://sdr.osmocom.org/trac/wiki/rtl-sdr)

### Building
This project requires the package [`github.com/bemasher/rtltcp`](http://godoc.org/github.com/bemasher/rtltcp), which provides a means of controlling and sampling from rtl-sdr dongles via the `rtl_tcp` tool. This package will be automatically downloaded and installed when getting rtlcap. The following command should be all that is required to install rtlcap:

	go get github.com/bemasher/rtlcap

This will produce the binary `$GOPATH/bin/rtlcap`. For convenience it's common to add `$GOPATH/bin` to the path.

### Usage
Available command-line flags are as follows:

```
Usage of rtlcap:
  -agcmode
    	enable/disable rtl agc
  -blocksize int
    	number of samples per block (default 4096)
  -bytes value
    	number of bytes to capture
  -centerfreq value
    	center frequency to receive on (default 100M)
  -directsampling
    	enable/disable direct sampling
  -duration duration
    	length of time to capture
  -freqcorrection int
    	frequency correction in ppm
  -gainbyindex uint
    	set gain by index
  -o string
    	filename to write samples to (default "/dev/null")
  -offsettuning
    	enable/disable offset tuning
  -quietsquelch
    	suppress log output messages for squelched blocks
  -rtlxtalfreq uint
    	set rtl xtal frequency
  -samplerate value
    	sample rate (default 2.4M)
  -server string
    	address or hostname of rtl_tcp instance (default "127.0.0.1:1234")
  -squelch float
    	minimum mean level a sample block must be to commit to disk
  -testmode
    	enable/disable test mode
  -tunergain float
    	set tuner gain in dB
  -tunergainmode
    	enable/disable tuner gain
  -tunerxtalfreq uint
    	set tuner xtal frequency
```

Running is as simple as starting an `rtl_tcp` instance and then starting rtlcap:

```bash
# Terminal A
$ rtl_tcp

# Terminal B
$ rtlcap
```

### Sensitivity
Keep in mind that if you are very close to the signal source it may be necessary to disable the tuner's AGC and set a fixed gain, this can be done using the following flags:

```
rtlcap -tunergainmode=false -tunergain=10.0
```

Or with `-gainbyindex`:

```
rtlcap -tunergainmode=false -gainbyindex=15
```

### Squelch
The `-squelch` flag is particularly useful for capturing intermittent signals. In order to use the squelch flag you need to know what the usual noise floor power level is. Once you've adjusted your sensitivity to produce a reasonable power level for the signal you're attempting to capture you can run rtlcap without specifying a filename with the `-o` flag to discard samples but display the minimum and maximum average power level for each block received over the last second:

```
rtlcap -centerfreq=912M -tunergainmode=false -gainbyindex=29
05:19:00.310460 Min: 0.000 Max: 0.670
05:19:01.334456 Min: 0.659 Max: 0.669
05:19:02.294447 Min: 0.659 Max: 0.668
05:19:03.318420 Min: 0.659 Max: 0.670
05:19:04.342942 Min: 0.660 Max: 0.671
05:19:05.302277 Min: 0.659 Max: 0.669
05:19:06.326606 Min: 0.659 Max: 0.669
05:19:07.287502 Min: 0.658 Max: 0.720
05:19:08.310351 Min: 0.659 Max: 0.717
05:19:09.334315 Min: 0.659 Max: 0.670
05:19:10.294277 Min: 0.658 Max: 0.669
05:19:11.318409 Min: 0.659 Max: 0.668
05:19:12.342544 Min: 0.659 Max: 0.670
05:19:13.302414 Min: 0.659 Max: 0.669
```

From the above output we can tell that the noise floor is approximately at 0.671, setting `-squelch=0.672` would discard any sample blocks whose average power level is below 0.672. Once you know a suitable squelch level, be sure to specify an output file with `-o` to capture the signal.

```
rtlcap -centerfreq=912M -tunergainmode=false -gainbyindex=29 -squelch=0.672 -bytes=25M
```

The above flags will capture 25MB worth of signal, keeping only blocks which satisfy the squelch level.

### License
The source of this project is licensed under Affero GPL v3.0. According to [http://choosealicense.com/licenses/agpl-3.0/](http://choosealicense.com/licenses/agpl-3.0/) you may:

#### Required:

 * **Disclose Source:** Source code must be made available when distributing the software. In the case of LGPL, the source for the library (and not the entire program) must be made available.
 * **License and copyright notice:** Include a copy of the license and copyright notice with the code.
 * **Network Use is Distribution:** Users who interact with the software via network are given the right to receive a copy of the corresponding source code.
 * **State Changes:** Indicate significant changes made to the code.

#### Permitted:

 * **Commercial Use:** This software and derivatives may be used for commercial purposes.
 * **Distribution:** You may distribute this software.
 * **Modification:** This software may be modified.
 * **Patent Grant:** This license provides an express grant of patent rights from the contributor to the recipient.
 * **Private Use:** You may use and modify the software without distributing it.

#### Forbidden:

 * **Hold Liable:** Software is provided without warranty and the software author/license owner cannot be held liable for damages.
 * **Sublicensing:** You may not grant a sublicense to modify and distribute this software to third parties not included in the license.

### Feedback
If you have any general questions or feedback leave a comment below. For bugs, feature suggestions and anything directly relating to the program itself, submit an issue in github or email me.
