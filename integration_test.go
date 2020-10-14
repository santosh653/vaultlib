package vaultlib

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
)

var vaultRoleID, vaultSecretID, noKVRoleID, noKVSecretID, longLivedRoleID, longLivedSecretID string

var vaultVersion string = *flag.String("vaultVersion", "1.0.1", "provide vault version to be tested against")

func TestMain(m *testing.M) {

	fmt.Println("Testing with Vault version", vaultVersion)
	fmt.Println("TestMain: Preparing Vault server")
	prepareVault()
	ret := m.Run()
	os.Exit(ret)
}

func execCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "VAULT_TOKEN=my-dev-root-vault-token")
	cmd.Env = append(cmd.Env, "VAULT_ADDR=http://localhost:8200")
	return cmd.Output()
}

func prepareVault() {
	err := startVault(vaultVersion)
	if err != nil {
		log.Fatalf("Error in initVaultDev.sh %v", err)
	}

	out, err := execCommand("./vault", "read", "-field=role_id", "auth/approle/role/my-role/role-id")
	if err != nil {
		log.Fatalf("error getting role id %v %v", err, out)
	}
	vaultRoleID = string(out)

	out, err = execCommand("./vault", "write", "-field=secret_id", "-f", "auth/approle/role/my-role/secret-id")
	if err != nil {
		log.Fatalf("error getting secret id %v", err)
	}
	vaultSecretID = string(out)

	out, err = execCommand("./vault", "read", "-field=role_id", "auth/approle/role/no-kv/role-id")
	if err != nil {
		log.Fatalf("error getting role id %v %v", err, out)
	}
	noKVRoleID = string(out)

	out, err = execCommand("./vault", "write", "-field=secret_id", "-f", "auth/approle/role/no-kv/secret-id")
	if err != nil {
		log.Fatalf("error getting secret id %v", err)
	}
	noKVSecretID = string(out)

	out, err = execCommand("./vault", "read", "-field=role_id", "auth/approle/role/long-lived/role-id")
	if err != nil {
		log.Fatalf("error getting role id %v %v", err, out)
	}
	longLivedRoleID = string(out)

	out, err = execCommand("./vault", "write", "-field=secret_id", "-f", "auth/approle/role/long-lived/secret-id")
	if err != nil {
		log.Fatalf("error getting secret id %v", err)
	}
	longLivedSecretID = string(out)

	os.Unsetenv("VAULT_TOKEN")
	fmt.Println("Vault initialized successfully")
}

func startVault(version string) error {
	cmd := exec.Command("bash", "./test-files/initVaultDev.sh", version)
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil

}
