package proxy

// CredentialStore 用于用户名密码认证
type CredentialStore interface {
	Valid(user, password string) bool
}

// StaticCredentials 使用内存实现用户名密码认证
type StaticCredentials map[string]string

func (s StaticCredentials) Valid(user, password string) bool {
	pass, ok := s[user]
	if !ok {
		return false
	}
	return password == pass
}
