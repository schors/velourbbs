/* crc16.c: the CRC-16 of XMODEM. */
#include <stdint.h>

/* crc16 gives the CRC-16/XMODEM of the n bytes of p. */
uint16_t crc16(const uint8_t *p, int n)
{
	uint16_t crc = 0;
	int i;

	while (n-- > 0) {
		crc ^= (uint16_t)*p++ << 8;
		for (i = 0; i < 8; i++) {
			if (crc & 0x8000)
				crc = (crc << 1) ^ 0x1021;	/* the polynomial of XMODEM */
			else
				crc <<= 1;
		}
	}
	return crc;
}

int main(void)
{
	static const uint8_t s[] = "123456789";
	return crc16(s, 9) == 0x31C3 ? 0 : 1;
}
