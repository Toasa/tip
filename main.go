package main

import (
	"os"
	"io/ioutil"
	"hash/crc32"
)

func main() {
	filename := "hello.txt"
	filedir := "test/"
	zip_filename := "hello.zip"
	z := newZipWriter(filename, filedir, zip_filename)
	z.Write()
}

type ZipWriter struct {
	filename string
	filedir string
	zip_filename string
	zip_data []byte

	local_file_header_len int
	file_data_len int
	central_dir_header_len int
	end_of_central_dir_record_len int
}

func newZipWriter(filename string, filedir string, zip_filename string) *ZipWriter {
	z := new(ZipWriter)
	z.filename = filename
	z.filedir = filedir
	z.zip_filename = zip_filename
	return z
}

func (z *ZipWriter) Write() {
	file, _ := os.Open(z.filedir + z.filename)
	data, _ := ioutil.ReadAll(file)

	z.writeLocalFileHeader(data)
	z.writeFileData(data)
	z.writeCentralDirectoryHeader(data)
	z.writeEndOfCentralDirectoryRecord()

	file, _ = os.Create(z.filedir + z.zip_filename)
	_, _ = file.Write(z.zip_data)
}

func (z *ZipWriter) writeLocalFileHeader(data []byte) {
	local_file_header_sign := []byte { 0x50, 0x4b, 0x03, 0x04 }

	// 解凍に必要な最低限のバージョン。
	// 0x0aは無圧縮。
	version_needed_to_extract := []byte{ 0x0a, 0x00 }
	general_purpose_bit_flag := []byte { 0x00, 0x00 }

	// 0 は無圧縮
	compression_method := []byte{ 0x00, 0x00 }

	// 時刻は0:0:0に決め打ち
	last_mod_file_time := []byte{ 0x00, 0x00 }
	// 日付は2020/8/31に決め打ち
	last_mod_file_date := []byte{ 0x1f, 0x51 }

	crc32 := u32_to_le_bytes(crc32.ChecksumIEEE(data))
	compressed_size := u32_to_le_bytes(uint32(len(data)))
	uncompressed_size := u32_to_le_bytes(uint32(len(data)))
	file_name := []byte(z.filename)
	file_name_len := u16_to_le_bytes(uint16(len(file_name)))
	extra_field_length := []byte{ 0x00, 0x00 }

	result := []byte{}
	result = append(result, local_file_header_sign...)
	result = append(result, version_needed_to_extract...)
	result = append(result, general_purpose_bit_flag...)
	result = append(result, compression_method...)
	result = append(result, last_mod_file_time...)
	result = append(result, last_mod_file_date...)
	result = append(result, crc32...)
	result = append(result, compressed_size...)
	result = append(result, uncompressed_size...)
	result = append(result, file_name_len...)
	result = append(result, extra_field_length...)
	result = append(result, file_name...)

	z.local_file_header_len = len(result)
	z.zip_data = result
}

func (z *ZipWriter) writeFileData(data []byte) {
	z.file_data_len = len(data)
	z.zip_data = append(z.zip_data, data...)
}

func (z *ZipWriter) writeCentralDirectoryHeader(data []byte) {
	signature := []byte{ 0x50, 0x4b, 0x01, 0x02 }

	// 上位ビットは属性情報の互換性を表す。0x03はUNIX。
	// 下位ビットはZIP仕様のバージョン
	version_made_by := []byte{ 0x00, 0x03 }
	version_needed_to_extract := []byte{ 0x0a, 0x00 }
	general_purpose_bit_flag := []byte { 0x00, 0x00 }

	// 0 は無圧縮
	compression_method := []byte{ 0x00, 0x00 }

	// 時刻は0:0:0に決め打ち
	last_mod_file_time := []byte{ 0x00, 0x00 }
	// 日付は2020/8/31に決め打ち
	last_mod_file_date := []byte{ 0x1f, 0x51 }

	crc32 := u32_to_le_bytes(crc32.ChecksumIEEE(data))
	compressed_size := u32_to_le_bytes(uint32(len(data)))
	uncompressed_size := u32_to_le_bytes(uint32(len(data)))
	file_name := []byte(z.filename)
	file_name_len := u16_to_le_bytes(uint16(len(file_name)))
	extra_field_length := []byte{ 0x00, 0x00 }
	file_comment_length := []byte{ 0x00, 0x00 }
	disk_number_start := []byte{ 0x00, 0x00 }
	internal_file_attr := []byte{ 0x00, 0x00 }
	// -rw-r--r--
	external_file_attr := []byte{ 0x00, 0x00, 0xa4, 0x81 }
	relative_offset_of_local_header := []byte{ 0x00, 0x00, 0x00, 0x00 }

	result := []byte{}
	result = append(result, signature...)
	result = append(result, version_made_by...)
	result = append(result, version_needed_to_extract...)
	result = append(result, general_purpose_bit_flag...)
	result = append(result, compression_method...)
	result = append(result, last_mod_file_time...)
	result = append(result, last_mod_file_date...)
	result = append(result, crc32...)
	result = append(result, compressed_size...)
	result = append(result, uncompressed_size...)
	result = append(result, file_name_len...)
	result = append(result, extra_field_length...)
	result = append(result, file_comment_length...)
	result = append(result, disk_number_start...)
	result = append(result, internal_file_attr...)
	result = append(result, external_file_attr...)
	result = append(result, relative_offset_of_local_header...)
	result = append(result, file_name...)

	z.central_dir_header_len = len(result)
	z.zip_data = append(z.zip_data, result...)
}

func (z *ZipWriter) writeEndOfCentralDirectoryRecord() {
	data := []byte {
		// end of central dir signature
		0x50, 0x4b, 0x05, 0x06,
		// number of this disk
		0x00, 0x00,
		// number of the disk with the start of the central directory
		0x00, 0x00,
		// total number of entries in the central directory on this disk
		0x01, 0x00,
		// total number of entries in the central directory
		0x01, 0x00,
	}

	size_of_central_directory := u32_to_le_bytes(uint32(z.central_dir_header_len))
	offset_of_start_of_central_directory := u32_to_le_bytes(uint32(z.local_file_header_len + z.file_data_len))
	comment_len := []byte{ 0x00, 0x00 }

	data = append(data, size_of_central_directory...)
	data = append(data, offset_of_start_of_central_directory...)
	data = append(data, comment_len...)

	z.zip_data = append(z.zip_data, data...)
}

func u16_to_le_bytes(n uint16) []byte {
	s := []byte{}
	for i := 0; i < 2; i++ {
		tmp := byte(n % 256)
		s = append(s, tmp)
		n /= 256
	}
	return s
}

func u32_to_le_bytes(n uint32) []byte {
	s := []byte{}
	for i := 0; i < 4; i++ {
		tmp := byte(n % 256)
		s = append(s, tmp)
		n /= 256
	}
	return s
}
