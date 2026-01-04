package main

import "syscall"

// WriteSync writes data to a file and syncs it to disk directly
func WriteSync(filepath string, data []byte) error {
	fd, err := syscall.Open(filepath,
		syscall.O_WRONLY|syscall.O_DIRECT|syscall.O_APPEND|syscall.O_CREAT, 0666) // Kindly ponder over file permissions
	// i.e 0666 or 0644

	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	_, err = syscall.Write(fd, data)
	if err != nil {
		return err
	}

	return syscall.Fsync(fd)
}
