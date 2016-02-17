#include <stdlib.h>
#include <stdio.h>

// Fill with a circular shading a specific area of the image
// the circle is described by a center (cx, cy) and internal and external radiuses
// 
// Fill with circular shading, from (sx, sy) to (ex, ey) in the rectangle (x1, y1)-(x2, y2) of buffer
void linearShading(unsigned char *buffer, int stride, int x1, int y1, int x2, int y2, double startCol, double endCol, double sx, double sy, double ex, double ey) {
	int width = x2 - x1;
	int height = y2 - y1;

//	int pixCount = width * height;

	// Coordinate distances
	double xd = ex - sx;
	double yd = ey - sy;
	// Something
	double c1 = xd*sx + yd*sy;
	double c2 = xd*ex + yd*ey;
	double cd = c2 - c1;

	#pragma omp parallel for collapse(2)
	for (int y = y1; y < y2; y++) {
		for (int x = x1; x < x2; x++) {
			// Convert coordinates to double
			const double xx = (double)x / (double)width;
			const double yy = (double)y / (double)height;

			const double c = xd*xx + yd*yy;
			double color;
			if (c <= c1)
				color = startCol;
			else if (c >= c2)
				color = endCol;
			else
				color = (startCol * (c2 - c) + endCol * (c - c1)) / cd;

			unsigned char *pix = &buffer[y*stride + x*4];
			pix[0] = pix[1] = pix[2] = (unsigned char)(color * 0xff);
			pix[3] = 0xff;
		}
	}
}

