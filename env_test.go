package gohera

import "testing"

func TestPickDeployEnv(t *testing.T) {
	cases := []struct {
		name      string
		flagVal   string
		flagSet   bool
		appEnv    string
		want      string
	}{
		{name: "default", want: DeployEnvDev},
		{name: "flag default unused when APP_ENV set", flagVal: DeployEnvDev, flagSet: false, appEnv: DeployEnvProd, want: DeployEnvProd},
		{name: "explicit flag wins", flagVal: DeployEnvTest, flagSet: true, appEnv: DeployEnvProd, want: DeployEnvTest},
		{name: "APP_ENV only", appEnv: DeployEnvPre, want: DeployEnvPre},
		{name: "empty APP_ENV falls to flag default", flagVal: DeployEnvDev, flagSet: false, appEnv: "", want: DeployEnvDev},
		{name: "explicit empty flag ignored", flagVal: "", flagSet: true, appEnv: DeployEnvProd, want: DeployEnvProd},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := pickDeployEnv(tc.flagVal, tc.flagSet, tc.appEnv); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestParseEnvFromAPP_ENV(t *testing.T) {
	t.Cleanup(func() {
		appEnv, appMode, appName = "", "", ""
	})
	t.Setenv("APP_ENV", "PROD")
	if err := parseEnv(pickDeployEnv(DeployEnvDev, false, "PROD")); err != nil {
		t.Fatal(err)
	}
	if GetEnv() != DeployEnvProd {
		t.Fatalf("GetEnv=%q", GetEnv())
	}
	if !IsProd() {
		t.Fatal("expected prod")
	}
}

func TestParseEnvRejectsUnknown(t *testing.T) {
	if err := parseEnv("staging"); err == nil {
		t.Fatal("expected invalid environment")
	}
}
