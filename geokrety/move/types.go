package move

func TypeName(typeID int16) string {
	switch typeID {
	case 0:
		return "Dropped"
	case 1:
		return "Grabbed"
	case 2:
		return "Commented"
	case 3:
		return "Seen"
	case 4:
		return "Archived"
	case 5:
		return "Dipped"
	default:
		return "Unknown"
	}
}
