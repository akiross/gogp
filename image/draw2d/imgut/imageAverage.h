#include <stdlib.h>
#include <stdio.h>

// Add every value in data into buffer
void imageAccumulate(float *buffer, unsigned char *data, int stride, int width, int height) {
	#pragma omp parallel for collapse(2)
	for (int y = 0; y < height; y++) {
		for (int x = 0; x < width; x++) {
			// Sum value to buffer
			unsigned char *pix = &data[y*stride + x*4];
			float *acc = &buffer[y * width * 4 + x * 4];
			for (int i = 0; i < 4; i++)
				acc[i] += pix[i];
		}
	}
}

void imageDivide(unsigned char *buffer, float *data, float divisor, int stride, int width, int height) {
	for (int y = 0; y < height; y++) {
		for (int x = 0; x < width; x++) {
			unsigned char *pix = &buffer[y*stride + x*4];
			float *dat = &data[y * width * 4 + x * 4];
			for (int i = 0; i < 4; i++)
				pix[i] = (unsigned char)(dat[i] / divisor);
		}
	}
}

