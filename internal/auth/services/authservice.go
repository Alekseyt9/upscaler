package services

type Claims struct {
	UserID       string
	Roles        []string
	Permissions  map[string][]string
	CustomClaims map[string]interface{}
	//jwt.StandardClaims
}

type AuthService interface {
	GenerateToken(userID string, roles []string, customClaims map[string]interface{}) (string, error)
	ValidateToken(token string) (bool, error)
	ExtractClaims(token string) (*Claims, error)
	GetPermissionsForAction(token string, action string) ([]string, error)
}
