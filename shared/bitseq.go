package shared

type BitSeq struct {
	Length int
	Data   [8]byte
}

func (bs BitSeq) Add_bit(bit bool) (res BitSeq) {
	copy(res.Data[:], bs.Data[:])
	if bit {
		res.Data[bs.Length/8] |= (1 << (bs.Length % 8))
	}

	res.Length = bs.Length + 1
	return
}
