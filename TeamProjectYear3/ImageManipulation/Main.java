import org.opencv.core.Mat;
import org.opencv.core.Size;
import org.opencv.highgui.Highgui;
import org.opencv.imgproc.Imgproc;

public class Main {
	static Mat src;
	static Mat src_gray;
	static Mat dst;
	static Mat detected_edges;

	int edgeThresh = 1;
	static int lowThreshold;
	final int max_lowThreshold = 100;
	static int ratio = 3;
	static int kernel_size = 3;

	// Note the following:
	// a. We establish a ratio of lower:upper threshold of 3:1 (with the
	// variable *ratio*)
	// b. We set the kernel size of :math:`3` (for the Sobel operations to be
	// performed internally by the Canny function)
	// c. We set a maximum value for the lower Threshold of :math:`100`.

	public static int main(String args[]) {
		// Load the image
		src = Highgui.imread(args[0]);
		if (!src.empty()) {
			return -1;
		}
		// Create a matrix of the same type and size
		dst.create(src.size(), src.type());

		// Convert to grey scale
		Imgproc.cvtColor(src, src_gray, Imgproc.COLOR_BGR2GRAY);

		// Reduce the noise with a kernel 3x3
		Imgproc.blur(src_gray, detected_edges, new Size(3, 3));

		// Call Canny detector
		Imgproc.Canny(detected_edges, detected_edges, lowThreshold,
				lowThreshold * ratio, kernel_size, false);

		// END OF CANNY EDGE DETECTION PART. THE IMAGE IS IN DETECTED_EDGES

		Imgproc.HoughLines(src, detected_edges, 0.5, 0.5, 1);

		// http://docs.opencv.org/doc/tutorials/imgproc/imgtrans/sobel_derivatives/sobel_derivatives.html

		return 0;
	}
}
