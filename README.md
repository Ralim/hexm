# hexm
CLI tool for working with intel hex and binary files.
Merge and convert, with some rebase support for bin files.
Roughly trying to avoid `objcopy` magic when working with split firmware images.

Possibly has bugs, but has reasonable coverage on the happy paths at the least.


## Features

* Splice hex files together
* Splice bin files together
* Convert hex file to bin
* Convert bin to hex file
* hex -> bin with user selectable starting point


### Examples

* Merge two hex files (with and without overlap)
* -> `hexm file1.hex file2.hex out.bin:0x100`
* Merge a bin and hex file
* Convert a hex file to a binary file (Optionally set base address of the output bin file)
* Truncate beginning of hex/bin file