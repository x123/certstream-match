/*
Copyright © 2024 x123 <x123@users.noreply.github.com>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Domain struct {
	gorm.Model
	Name string
}

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "interact with the database",
	Long:  `interact with the database`,
	Run:   initDb,
}

func initDb(cmd *cobra.Command, args []string) {
	fmt.Println("initDb")
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Domain{})
	db.Create(
		&Domain{
			Name: "example.com",
		})
	var domain Domain
	// db.First(&domain, 1)
	db.First(&domain, "name = ?", "example.com")
	fmt.Printf("%s\n", domain.Name)
}

func init() {
	rootCmd.AddCommand(dbCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
