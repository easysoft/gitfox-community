package lock

func formatKey(app, ns, key string) string {
	return app + ":" + ns + ":" + key
}
