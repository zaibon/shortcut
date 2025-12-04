package components

import "strings"

func GetSourceIcon(source string) (icon, bg, text string) {
	s := strings.ToLower(source)
	switch {
	case strings.Contains(s, "google"):
		return "fab fa-google", "bg-red-50", "text-red-600"
	case strings.Contains(s, "twitter") || strings.Contains(s, "t.co") || strings.Contains(s, "x.com"):
		return "fab fa-twitter", "bg-blue-50", "text-blue-400"
	case strings.Contains(s, "facebook"):
		return "fab fa-facebook", "bg-blue-50", "text-blue-700"
	case strings.Contains(s, "linkedin"):
		return "fab fa-linkedin", "bg-blue-50", "text-blue-800"
	case strings.Contains(s, "reddit"):
		return "fab fa-reddit-alien", "bg-orange-50", "text-orange-600"
	case strings.Contains(s, "instagram"):
		return "fab fa-instagram", "bg-pink-50", "text-pink-600"
	case strings.Contains(s, "youtube"):
		return "fab fa-youtube", "bg-red-50", "text-red-600"
	case strings.Contains(s, "github"):
		return "fab fa-github", "bg-gray-50", "text-gray-800"
	case s == "direct" || s == "":
		return "fas fa-link", "bg-gray-50", "text-gray-600"
	default:
		return "fas fa-globe", "bg-indigo-50", "text-indigo-600"
	}
}
