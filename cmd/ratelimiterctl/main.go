package main

func main() {
	rootCmd := rootCmd()
	rootCmd.AddCommand(registCmd())
	rootCmd.AddCommand(deleteCmd())
	rootCmd.AddCommand(rcListCmd())
	rootCmd.Execute()
}
