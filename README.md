# OpenTelemetry Profile Symbolization Processor
This OpenTelemetry Collector Processor performs the symbolization of native profiles.
It relies on [blazesym-c](https://github.com/libbpf/blazesym/tree/main/capi) to perform address symbolization.

# Dependencies
- [blazesym-c](https://github.com/libbpf/blazesym/tree/main/capi): you must build the blazesym-c library `libblazesym_c.a` and install the library and the header `blazesym.h` in a location where the C compiler can find them (typically `/usr/local/lib` `for the library and /usr/local/include` for the header)