package auth

func generateToken(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func toInt32Slice(roles []domain.Role) []int32 {
	out := make([]int32, len(roles))
	for i, r := range roles {
		out[i] = int32(r)
	}
	return out
}