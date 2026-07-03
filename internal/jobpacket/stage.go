package jobpacket

func ValidStage(stage int) bool {
	return stage >= 0 && stage <= 3
}
