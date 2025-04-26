package cmd

import "github.com/tanq16/anbu/utils"

func playground() {
	utils.PrintSuccess("Playground success test")
	utils.PrintSuccess2("Playground success test 2")
	utils.PrintError("Playground error test")
	utils.PrintWarning("Playground warning test")
	utils.PrintInfo("Playground info test")
	utils.PrintDebug("Playground debug test")
	utils.PrintDetail("Playground detail test")
	utils.PrintHeader("Playground header test")
	utils.PrintStream("Playground stream test")
	// om := utils.GetDefaultManager()
	// funcID := om.Register("playground")
	// om.SetMessage(funcID, "Playground started")
	// time.Sleep(4 * time.Second)
	// om.SetMessage(funcID, "Playground done executing")
	// time.Sleep(2 * time.Second)
	// om.Complete(funcID, "Playground completed")
	// time.Sleep(2 * time.Second)
}
