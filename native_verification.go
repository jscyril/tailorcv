package main

import (
	"fmt"

	"github.com/jscyril/tailorcv/internal/credentials"
)

const verifyNativeCredentialsFlag = "--verify-native-credentials"
const verifyPackagedWorkflowFlag = "--verify-packaged-workflow"

func runVerificationCommand(arguments []string) (bool, error) {
	if len(arguments) != 1 {
		return false, nil
	}
	switch arguments[0] {
	case verifyNativeCredentialsFlag:
		if err := credentials.VerifyNativeLifecycle(); err != nil {
			return true, err
		}
		fmt.Println("TailorCV native credential verification passed.")
		return true, nil
	case verifyPackagedWorkflowFlag:
		if err := verifyPackagedWorkflow(); err != nil {
			return true, err
		}
		fmt.Println("TailorCV packaged workflow verification passed.")
		return true, nil
	default:
		return false, nil
	}
}
