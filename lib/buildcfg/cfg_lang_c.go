package buildcfg

func loadCConfig(remote string, c *Config) {
	c.Addons.Apt.Packages = append(c.Addons.Apt.Packages, "build-essential", "cmake", "libssl-dev", "python-dev", "libffi-dev")
	c.DockerImage = "debian"
}
