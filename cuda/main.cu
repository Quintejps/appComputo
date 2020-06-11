#include "cuda_runtime.h"
#include "device_launch_parameters.h"
#include <opencv2/core/core.hpp>
#include <opencv2/highgui/highgui.hpp>

using namespace cv;
using namespace std;

__global__ void BinaryCUDA(unsigned char* Image, int Channels, int umbral) {
	int x = blockIdx.x;
    int y = blockIdx.y;
	int idx = (x + y * gridDim.x) * Channels;
	for (int i = 0; i < Channels; i++) {
		Image[idx + i] = 255 - Image[idx + i];
		if (Image[idx+i] > umbral) {
			Image[idx + i] = 255;
		}
		else {
			Image[idx + i] = 0;
		}
	}
}

void ImageBinary(unsigned char* Input_Image, int Height, int Width, int Channels, int umbral) {
	unsigned char* Dev_Input_Image = NULL;

	//allocate the memory in gpu
	cudaMalloc((void**)&Dev_Input_Image, Height * Width * Channels);

	//copy data from CPU to GPU
	cudaMemcpy(Dev_Input_Image, Input_Image, Height * Width * Channels, cudaMemcpyHostToDevice);

	dim3 Grid_Image(Width, Height);
	BinaryCUDA<<<Grid_Image, 8>>>(Dev_Input_Image, Channels, umbral);

	//copy processed data back to cpu from gpu
	cudaMemcpy(Input_Image, Dev_Input_Image, Height * Width * Channels, cudaMemcpyDeviceToHost);

	//free gpu mempry
	cudaFree(Dev_Input_Image);
}

__global__ void BinaryInvCUDA(unsigned char* Image, int Channels, int umbral) {
	int x = blockIdx.x;
	int y = blockIdx.y;
	int idx = (x + y * gridDim.x) * Channels;

	for (int i = 0; i < Channels; i++) {
		Image[idx + i] = 255 - Image[idx + i];
		if (Image[idx + i] > umbral) {
			Image[idx + i] = 0;
		}
		else {
			Image[idx + i] = 255;
		}
	}
}

void ImageBinaryInv(unsigned char* Input_Image, int Height, int Width, int Channels, int umbral) {
	unsigned char* Dev_Input_Image = NULL;

	//allocate the memory in gpu
	cudaMalloc((void**)&Dev_Input_Image, Height * Width * Channels);

	//copy data from CPU to GPU
	cudaMemcpy(Dev_Input_Image, Input_Image, Height * Width * Channels, cudaMemcpyHostToDevice);

	dim3 Grid_Image(Width, Height);
	BinaryInvCUDA <<<Grid_Image, 8>>>(Dev_Input_Image, Channels, umbral);

	//copy processed data back to cpu from gpu
	cudaMemcpy(Input_Image, Dev_Input_Image, Height * Width * Channels, cudaMemcpyDeviceToHost);

	//free gpu mempry
	cudaFree(Dev_Input_Image);
}

__global__ void TruncCUDA(unsigned char* Image, int Channels, int umbral) {
	int x = blockIdx.x;
	int y = blockIdx.y;
	int idx = (x + y * gridDim.x) * Channels;

	for (int i = 0; i < Channels; i++) {
		Image[idx + i] = 255 - Image[idx + i];
		if (Image[idx + i] > umbral) {
			Image[idx + i] = Image[idx + i];
		}
		else {
			Image[idx + i] = 255;
		}
	}
}

void ImageTrunc(unsigned char* Input_Image, int Height, int Width, int Channels, int umbral) {
	unsigned char* Dev_Input_Image = NULL;

	//allocate the memory in gpu
	cudaMalloc((void**)&Dev_Input_Image, Height * Width * Channels);

	//copy data from CPU to GPU
	cudaMemcpy(Dev_Input_Image, Input_Image, Height * Width * Channels, cudaMemcpyHostToDevice);

	dim3 Grid_Image(Width, Height);
	TruncCUDA<<<Grid_Image, 8>>>(Dev_Input_Image, Channels, umbral);

	//copy processed data back to cpu from gpu
	cudaMemcpy(Input_Image, Dev_Input_Image, Height * Width * Channels, cudaMemcpyDeviceToHost);

	//free gpu mempry
	cudaFree(Dev_Input_Image);
}

__global__ void TozeroCUDA(unsigned char* Image, int Channels, int umbral) {
	int x = blockIdx.x;
	int y = blockIdx.y;
	int idx = (x + y * gridDim.x) * Channels;

	for (int i = 0; i < Channels; i++) {
		if (Image[idx + i] > umbral) {
			Image[idx + i] = 0;
		}
		else {
			Image[idx + i] = Image[idx + i];
		}
	}
}

void ImageTozero(unsigned char* Input_Image, int Height, int Width, int Channels, int umbral) {
	unsigned char* Dev_Input_Image = NULL;

	//allocate the memory in gpu
	cudaMalloc((void**)&Dev_Input_Image, Height * Width * Channels);

	//copy data from CPU to GPU
	cudaMemcpy(Dev_Input_Image, Input_Image, Height * Width * Channels, cudaMemcpyHostToDevice);

	dim3 Grid_Image(Width, Height);
	TruncCUDA<<<Grid_Image, 8>>>(Dev_Input_Image, Channels, umbral);

	//copy processed data back to cpu from gpu
	cudaMemcpy(Input_Image, Dev_Input_Image, Height * Width * Channels, cudaMemcpyDeviceToHost);

	//free gpu mempry
	cudaFree(Dev_Input_Image);
}


__global__ void TozeroInvCUDA(unsigned char* Image, int Channels, int umbral) {
	int x = blockIdx.x;
	int y = blockIdx.y;
	int idx = (x+y*gridDim.x) * Channels;

	for (int i = 0; i < Channels; i++) {
		if (Image[idx + i] > umbral) {
			Image[idx + i] = Image[idx + i];
		}
		else {
			Image[idx + i] = 0;
		}
	}
}

void ImageTozeroInv(unsigned char* Input_Image, int Height, int Width, int Channels, int umbral) {
	unsigned char* Dev_Input_Image = NULL;


	//allocate the memory in gpu
	cudaMalloc((void**)&Dev_Input_Image, Height * Width * Channels);

	//copy data from CPU to GPU
	cudaMemcpy(Dev_Input_Image, Input_Image, Height * Width * Channels, cudaMemcpyHostToDevice);

	dim3 Grid_Image(Width, Height);
	TozeroInvCUDA<<<Grid_Image, 8>>>(Dev_Input_Image, Channels, umbral);

	//copy processed data back to cpu from gpu
	cudaMemcpy(Input_Image, Dev_Input_Image, Height * Width * Channels, cudaMemcpyDeviceToHost);

	//free gpu mempry
	cudaFree(Dev_Input_Image);
}

int main(int argc, char* argv[])
{
	Mat image = imread(argv[1], CV_LOAD_IMAGE_COLOR);
	int umbral = 112;

	if (!image.data)
	{
		printf("Could not open or find the image\n");
		return -1;
    }
    Mat output_image = image.clone();

    string option = argv[2];

	int o = 0;

	if (option == "binary") {
		o = 1;
	}
	else if (option == "binaryInv") {
		o = 2;
	}
	else if (option == "trunc") {
		o = 3;
	}
	else if (option == "tozero") {
		o = 4;
	}
	else if (option == "tozeroInv") {
		o = 5;
	}
	printf("\nImage processing starting\n");
	switch (o) {
		case 1:
			ImageBinary(output_image.data, output_image.cols, output_image.rows, output_image.channels(), umbral);
			break;
		case 2:
			ImageBinaryInv(output_image.data, output_image.cols, output_image.rows, output_image.channels(), umbral);
			break;
		case 3:
			ImageTrunc(output_image.data, output_image.cols, output_image.rows, output_image.channels(), umbral);
			break;
		case 4:
			ImageTozero(output_image.data, output_image.cols, output_image.rows, output_image.channels(), umbral);
			break;
		case 5:
			ImageTozeroInv(output_image.data, output_image.cols, output_image.rows, output_image.channels(), umbral);
			break;
	}

    imwrite(argv[1], output_image);
	printf("\nImage processing done\n");
}