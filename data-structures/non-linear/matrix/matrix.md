# MATRIX

Matrix is a two-dimensional data structure that stores data in a tabular form. It is a collection of elements arranged in rows and columns. Each element is accessed by its row and column number. The number of rows and columns in a matrix is called its order. The order of a matrix is represented by the tuple (m, n) where m is the number of rows and n is the number of columns.

## Matrix Representation

A matrix can be represented in two ways:

- Using a two-dimensional array
- Using a one-dimensional array

### Using a Two-Dimensional Array

A two-dimensional array can be used to represent a matrix. The number of rows and columns in the matrix is equal to the number of rows and columns in the two-dimensional array. The element at the i<sup>th</sup> row and j<sup>th</sup> column of the matrix is stored at the i<sup>th</sup> row and j<sup>th</sup> column of the two-dimensional array.

### Using a One-Dimensional Array

A one-dimensional array can also be used to represent a matrix. The number of rows and columns in the matrix is equal to the number of rows and columns in the two-dimensional array. The element at the i<sup>th</sup> row and j<sup>th</sup> column of the matrix is stored at the (i * n + j)<sup>th</sup> element of the one-dimensional array where n is the number of columns in the matrix.

## Matrix Operations

### Addition

The sum of two matrices A and B is a matrix C such that C<sub>i,j</sub> = A<sub>i,j</sub> + B<sub>i,j</sub> where A<sub>i,j</sub> and B<sub>i,j</sub> are the elements of the matrices A and B at the i<sup>th</sup> row and j<sup>th</sup> column respectively.

### Subtraction

The difference of two matrices A and B is a matrix C such that C<sub>i,j</sub> = A<sub>i,j</sub> - B<sub>i,j</sub> where A<sub>i,j</sub> and B<sub>i,j</sub> are the elements of the matrices A and B at the i<sup>th</sup> row and j<sup>th</sup> column respectively.

### Multiplication

The product of two matrices A and B is a matrix C such that C<sub>i,j</sub> = A<sub>i,1</sub> *B<sub>1,j</sub> + A<sub>i,2</sub>* B<sub>2,j</sub> + ... + A<sub>i,n</sub> * B<sub>n,j</sub> where A<sub>i,j</sub> and B<sub>i,j</sub> are the elements of the matrices A and B at the i<sup>th</sup> row and j<sup>th</sup> column respectively.

### Transpose

The transpose of a matrix A is a matrix B such that B<sub>i,j</sub> = A<sub>j,i</sub> where A<sub>i,j</sub> and B<sub>i,j</sub> are the elements of the matrices A and B at the i<sup>th</sup> row and j<sup>th</sup> column respectively.

### Determinant

The determinant of a matrix A is a scalar value that is calculated using the elements of the matrix. The determinant of a matrix is denoted by det(A).

### Inverse

The inverse of a matrix A is a matrix B such that A *B = B* A = I where I is the identity matrix. The inverse of a matrix is denoted by A<sup>-1</sup>.

## Matrix Applications

### Image Processing

Matrices are used to represent images. The elements of the matrix represent the intensity of the pixels in the image. The intensity of the pixel at the i<sup>th</sup> row and j<sup>th</sup> column is stored at the i<sup>th</sup> row and j<sup>th</sup> column of the matrix.
