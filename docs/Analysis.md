# Analysis of the Panasonic Inverter IR Controller A75C3115

## Examples of reverse engineering of Panasonic Inverter Remote Controllers

<https://www.analysir.com/blog/2014/12/27/reverse-engineering-panasonic-ac-infrared-protocol/>

<https://www.instructables.com/Reverse-engineering-of-an-Air-Conditioning-control/>

## Interpreting the raw LIRC data

The raw data provided by the LIRC kernel module consists of unsigned 32-bit integers. All data read from /dev/lirc-rc is LittleEndian, therefore each group of four bytes (32 bits) start with the least significant bit and end with the most significant bit.
The conversion to unsigned 32-bit integers must use LittleEndian conversion.

The unsigned integers are in a format known as LIRC Mode2. These represent the pulses and spaces detected by the IR receiver. Each integer is the duration of either a pulse or a space. The duration is never exact, so the data needs to be "cleaned" (rounded to the expected values). The cleaned-up LIRC Mode2 integers can then be interpreted as bits. It's actually the spaces that represent the bits, using two different durations, while pulses are all the same duration.

According to the previous analysis linked above, the data sent by the remote control consists of two frames. The first frame is constant, while the second contains the configuration. In these previous analysis, the authors have chosen to append each bit to the previous bits, and view them as bytes. This leads to some problems later when interpreting the bits, because configuration fields do not use 8 or 16 bits - instead they can use e.g. 5 or 11 bits, which may be split across byte boundaries. In addition, the bits are actually sent in LitleEndian (the first received bit is the least significant), which means the bit order would also need to be reversed to get the correct values.

Instead, I have chosen view to the received bits in each frame as a stream of bits, with the least significant bit first and the most significant bit last. Each frame is stored in a BigInt. The first received bit is saved at index 0 in the BigInt, the second at index 1 in the BigInt, and so on. This reverses the bit order directly, so that we end up with bits saved in normal BigEndian representation. To get the value of a field, we specify field's first bit (the index of the least significant bit) and number of bits (to the left of the first bit), and extract those bits with simple bit operations on the BigInt.
