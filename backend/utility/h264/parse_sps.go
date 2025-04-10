package h264

import (
	"fmt"
	"math"
)

// SPSInfo 包含 H.264 序列参数集的关键信息
type SPSInfo struct {
	ProfileIdc                uint
	ConstraintSetFlags        uint
	LevelIdc                  uint
	PicWidthInMbsMinus1       uint
	PicHeightInMapUnitsMinus1 uint
	FrameMbsOnlyFlag          uint
	FrameCropLeftOffset       uint
	FrameCropRightOffset      uint
	FrameCropTopOffset        uint
	FrameCropBottomOffset     uint
	Sar                       [2]uint // Sample Aspect Ratio
}

// ParseSPS 解析 H.264 SPS (序列参数集) 数据
// 输入参数是包含 NAL 单元的完整数据
func ParseSPS(data []byte) (*SPSInfo, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("无效的 SPS 数据长度")
	}

	// 去除防竞争字节
	cleanData := deEmulationPrevention(data[4:]) // 跳过 NAL header
	var startBit uint = 0
	nLen := uint(len(cleanData))

	info := &SPSInfo{}

	// 跳过 forbidden_zero_bit, nal_ref_idc
	readBits(3, cleanData, &startBit)

	// 检查 NAL unit type 是否为 SPS (7)
	nalUnitType := readBits(5, cleanData, &startBit)
	if nalUnitType != 7 {
		return nil, fmt.Errorf("不是 SPS NAL 单元")
	}

	info.ProfileIdc = readBits(8, cleanData, &startBit)
	info.ConstraintSetFlags = readBits(8, cleanData, &startBit)
	info.LevelIdc = readBits(8, cleanData, &startBit)

	// 跳过 seq_parameter_set_id
	readUe(cleanData, nLen, &startBit)

	// 处理高级配置
	if info.ProfileIdc == 100 || info.ProfileIdc == 110 ||
		info.ProfileIdc == 122 || info.ProfileIdc == 144 {
		chromaFormatIdc := readUe(cleanData, nLen, &startBit)
		if chromaFormatIdc == 3 {
			readBits(1, cleanData, &startBit) // residual_colour_transform_flag
		}
		readUe(cleanData, nLen, &startBit) // bit_depth_luma_minus8
		readUe(cleanData, nLen, &startBit) // bit_depth_chroma_minus8
		readBits(1, cleanData, &startBit)  // qpprime_y_zero_transform_bypass_flag

		seqScalingMatrixPresentFlag := readBits(1, cleanData, &startBit)
		if seqScalingMatrixPresentFlag > 0 {
			for i := 0; i < 8; i++ {
				if readBits(1, cleanData, &startBit) > 0 {
					// 跳过缩放列表
					for j := 0; j < 64; j++ {
						readUe(cleanData, nLen, &startBit)
					}
				}
			}
		}
	}

	readUe(cleanData, nLen, &startBit) // log2_max_frame_num_minus4
	picOrderCntType := readUe(cleanData, nLen, &startBit)

	if picOrderCntType == 0 {
		readUe(cleanData, nLen, &startBit)
	} else if picOrderCntType == 1 {
		readBits(1, cleanData, &startBit)
		readSe(cleanData, nLen, &startBit)
		readSe(cleanData, nLen, &startBit)
		numRefFramesInPicOrderCntCycle := readUe(cleanData, nLen, &startBit)
		for i := uint(0); i < numRefFramesInPicOrderCntCycle; i++ {
			readSe(cleanData, nLen, &startBit)
		}
	}

	readUe(cleanData, nLen, &startBit) // max_num_ref_frames
	readBits(1, cleanData, &startBit)  // gaps_in_frame_num_value_allowed_flag

	info.PicWidthInMbsMinus1 = readUe(cleanData, nLen, &startBit)
	info.PicHeightInMapUnitsMinus1 = readUe(cleanData, nLen, &startBit)

	info.FrameMbsOnlyFlag = readBits(1, cleanData, &startBit)
	if info.FrameMbsOnlyFlag == 0 {
		readBits(1, cleanData, &startBit) // mb_adaptive_frame_field_flag
	}

	readBits(1, cleanData, &startBit) // direct_8x8_inference_flag

	frameCroppingFlag := readBits(1, cleanData, &startBit)
	if frameCroppingFlag > 0 {
		info.FrameCropLeftOffset = readUe(cleanData, nLen, &startBit)
		info.FrameCropRightOffset = readUe(cleanData, nLen, &startBit)
		info.FrameCropTopOffset = readUe(cleanData, nLen, &startBit)
		info.FrameCropBottomOffset = readUe(cleanData, nLen, &startBit)
	}

	vuiParametersPresentFlag := readBits(1, cleanData, &startBit)
	if vuiParametersPresentFlag > 0 {
		aspectRatioInfoPresentFlag := readBits(1, cleanData, &startBit)
		if aspectRatioInfoPresentFlag > 0 {
			aspectRatioIdc := readBits(8, cleanData, &startBit)
			if aspectRatioIdc == 255 { // Extended_SAR
				info.Sar[0] = readBits(16, cleanData, &startBit) // sar_width
				info.Sar[1] = readBits(16, cleanData, &startBit) // sar_height
			} else if aspectRatioIdc < 17 {
				// 使用预定义的 SAR 值
				info.Sar = getPredefinedSAR(aspectRatioIdc)
			}
		}
	}

	return info, nil
}

// 获取预定义的 SAR 值
func getPredefinedSAR(aspectRatioIdc uint) [2]uint {
	// 根据 H.264 标准定义的预设 SAR 值
	predefinedSAR := [][2]uint{
		{0, 0},    // Unspecified
		{1, 1},    // 1:1
		{12, 11},  // 12:11
		{10, 11},  // 10:11
		{16, 11},  // 16:11
		{40, 33},  // 40:33
		{24, 11},  // 24:11
		{20, 11},  // 20:11
		{32, 11},  // 32:11
		{80, 33},  // 80:33
		{18, 11},  // 18:11
		{15, 11},  // 15:11
		{64, 33},  // 64:33
		{160, 99}, // 160:99
		{4, 3},    // 4:3
		{3, 2},    // 3:2
		{2, 1},    // 2:1
	}

	if aspectRatioIdc < uint(len(predefinedSAR)) {
		return predefinedSAR[aspectRatioIdc]
	}
	return [2]uint{0, 0}
}

// 用于解析 Exp-Golomb 编码的工具函数
func readUe(data []byte, nLen uint, startBit *uint) uint {
	// 计算前导零的个数
	nZeroNum := 0
	for *startBit < nLen*8 {
		if (data[*startBit/8] & (0x80 >> (*startBit % 8))) > 0 {
			break
		}
		nZeroNum++
		*startBit++
	}
	*startBit++

	// 计算结果
	dwRet := uint(0)
	for i := 0; i < nZeroNum; i++ {
		dwRet <<= 1
		if (data[*startBit/8] & (0x80 >> (*startBit % 8))) > 0 {
			dwRet += 1
		}
		*startBit++
	}
	return (1 << uint(nZeroNum)) - 1 + dwRet
}

// 读取有符号的 Exp-Golomb 编码
func readSe(data []byte, nLen uint, startBit *uint) int {
	ueVal := readUe(data, nLen, startBit)
	k := int64(ueVal)
	nValue := int(math.Ceil(float64(k) / 2))
	if ueVal%2 == 0 {
		nValue = -nValue
	}
	return nValue
}

// 读取指定位数的比特
func readBits(bitCount uint, data []byte, startBit *uint) uint {
	dwRet := uint(0)
	for i := uint(0); i < bitCount; i++ {
		dwRet <<= 1
		if (data[*startBit/8] & (0x80 >> (*startBit % 8))) > 0 {
			dwRet += 1
		}
		*startBit++
	}
	return dwRet
}

// 处理防竞争机制
func deEmulationPrevention(data []byte) []byte {
	size := len(data)
	tmpData := make([]byte, size)
	copy(tmpData, data)

	j := 0
	for i := 0; i < size-2; i++ {
		if i+2 < size &&
			tmpData[i] == 0x00 &&
			tmpData[i+1] == 0x00 &&
			tmpData[i+2] == 0x03 {
			tmpData[j] = tmpData[i]
			tmpData[j+1] = tmpData[i+1]
			i += 2
			j += 2
		} else {
			tmpData[j] = tmpData[i]
			j++
		}
	}
	// 复制剩余的字节
	for i := size - 2; i < size; i++ {
		if j < size {
			tmpData[j] = data[i]
			j++
		}
	}
	return tmpData[:j]
}
