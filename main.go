package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	pokecache "github.com/TheOTG/pokedexcli/internal"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, string) error
}

type config struct {
	next     string
	previous string
}

func cleanInput(text string) []string {
	if text == "" {
		return []string{}
	}

	cleaned := strings.Fields(strings.ToLower(text))

	return cleaned
}

func commandMap(cfg *config, s string) error {
	baseURL := cfg.next
	v, ok := cache.Get(baseURL)
	var locationAreas LocationAreaData
	if ok {
		err := json.Unmarshal(v, &locationAreas)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(baseURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &locationAreas)
		if err != nil {
			return err
		}

		cache.Add(baseURL, data)
	}

	if locationAreas.Next != "null" {
		cfg.next = locationAreas.Next
	}
	if locationAreas.Previous != "null" {
		cfg.previous = locationAreas.Previous
	}

	for _, locationArea := range locationAreas.Results {
		fmt.Println(locationArea.Name)
	}

	return nil
}

func commandMapBack(cfg *config, s string) error {
	if cfg.previous == "" {
		fmt.Println("No previous locations")
		return nil
	}

	baseURL := cfg.previous
	v, ok := cache.Get(baseURL)
	var locationAreas LocationAreaData
	if ok {
		err := json.Unmarshal(v, &locationAreas)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(baseURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &locationAreas)
		if err != nil {
			return err
		}

		cache.Add(baseURL, data)
	}

	if locationAreas.Next != "null" {
		cfg.next = locationAreas.Next
	}
	if locationAreas.Previous != "null" {
		cfg.previous = locationAreas.Previous
	}

	for _, locationArea := range locationAreas.Results {
		fmt.Println(locationArea.Name)
	}

	return nil
}

func commandExplore(cfg *config, loc string) error {
	if loc == "" {
		fmt.Println("Missing location, please try again")
		return nil
	}
	fmt.Printf("Exploring %s...\n", loc)
	baseURL := "https://pokeapi.co/api/v2/location-area/"
	fullURL := baseURL + loc
	v, ok := cache.Get(fullURL)
	var exploreData ExploreLocation
	if ok {
		err := json.Unmarshal(v, &exploreData)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(fullURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &exploreData)
		if err != nil {
			return err
		}

		cache.Add(fullURL, data)
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range exploreData.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandCatch(cfg *config, pokemon string) error {
	if pokemon == "" {
		fmt.Println("Missing location, please try again")
		return nil
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon)
	baseURL := "https://pokeapi.co/api/v2/pokemon/"
	fullURL := baseURL + pokemon
	v, ok := cache.Get(fullURL)
	var pokemonData Pokemon
	if ok {
		err := json.Unmarshal(v, &pokemonData)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(fullURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &pokemonData)
		if err != nil {
			return err
		}

		cache.Add(fullURL, data)
	}

	check, ok := caughtPokemons[pokemonData.Name]
	if ok {
		fmt.Printf("You already have %s\n", check.Name)
		return nil
	}

	catchRate := 90 - (pokemonData.BaseExp / 10)
	num := rand.Intn(100)

	if num <= catchRate {
		fmt.Printf("%s was caught!\n", pokemonData.Name)
		caughtPokemons[pokemonData.Name] = pokemonData
	} else {
		fmt.Printf("%s escaped!\n", pokemonData.Name)
	}

	return nil
}

func commandInspect(cfg *config, pokemon string) error {
	pokemonData, ok := caughtPokemons[pokemon]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemonData.Name)
	fmt.Printf("Height: %d\n", pokemonData.Height)
	fmt.Printf("Weight: %d\n", pokemonData.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemonData.Stats {
		fmt.Printf(" -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, pokemonType := range pokemonData.Types {
		fmt.Printf(" - %s\n", pokemonType.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *config, s string) error {
	if len(caughtPokemons) == 0 {
		fmt.Println("You have not caught any Pokemon yet")
		return nil
	}

	fmt.Println("Your Pokedex:")
	for k, _ := range caughtPokemons {
		fmt.Printf(" - %s\n", k)
	}

	return nil
}

func commandHelp(cfg *config, loc string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, v := range commands {
		fmt.Printf("%s: %s\n", v.name, v.description)
	}

	return nil
}

func commandExit(cfg *config, loc string) error {
	fmt.Print("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}

var commands map[string]cliCommand
var caughtPokemons = map[string]Pokemon{}
var cfg config
var cache = pokecache.NewCache(5 * time.Second)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	commands = map[string]cliCommand{
		"map": {
			name:        "map",
			description: "Shows the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Shows the previous 20 location areas",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "List of all Pokemon in location area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect your caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all your caught Pokemon",
			callback:    commandPokedex,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}

	cfg.next = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		text := scanner.Text()
		cleaned := cleanInput(text)
		command, ok := commands[cleaned[0]]
		if !ok {
			fmt.Print("Unknown command\n")
			continue
		}
		var secondary string
		if len(cleaned) > 1 {
			secondary = cleaned[1]
		}
		err := command.callback(&cfg, secondary)
		if err != nil {
			log.Fatal(err)
		}
	}
}
