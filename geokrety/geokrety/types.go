package geokrety

func TypeName(typeID int16) string {
	switch typeID {
	case 0:
		return "Traditional"
	case 1:
		return "Book/CD/DVD..."
	case 2:
		return "Human/Pet"
	case 3:
		return "Coin"
	case 4:
		return "KretyPost"
	case 5:
		return "Pebble"
	case 6:
		return "Car"
	case 7:
		return "Playing card"
	case 8:
		return "Dog tag/pet"
	case 9:
		return "Jigsaw part"
	case 10:
		return "Hidden GeoKret"
	default:
		return "Unknown"
	}
}
