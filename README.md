# beacon.client

The [golang] client (device) application for the beacon platform.

**Requirements**

- [libusb] - native c library that is relied on by the blink1 [go library][blink-lib]

**Compiling on Mac**

Because libusb and libusb-compat are typically installed into `/usr/local`, compiling on mac osx can sometimes produce an `unable to find <usb.h>` error when compiling the application. This can be avoided by adding the following flags before compilation:

```
CGO_CFLAGS=-I/usr/local/include CGO_LDFLAGS=-L/usr/local/lib make
```

[golang]: https://golang.org
[libusb]: https://github.com/libusb/libusb
[blink-lib]: https://github.com/hink/go-blink1
