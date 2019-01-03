package rest

import (
	"hlc/app/models"
	"sync"
)

func merge(s []models.Account, middle int, a models.Account) {
	helper := make([]models.Account, len(s))
	copy(helper, s)

	helperLeft := 0
	helperRight := middle
	current := 0
	high := len(s) - 1

	for helperLeft <= middle-1 && helperRight <= high {
		if a.CheckSimilarity(helper[helperLeft]) >= a.CheckSimilarity(helper[helperRight]) {
			s[current] = helper[helperLeft]
			helperLeft++
		} else {
			s[current] = helper[helperRight]
			helperRight++
		}
		current++
	}

	for helperLeft <= middle-1 {
		s[current] = helper[helperLeft]
		current++
		helperLeft++
	}
}

func parallelMergeSort(s []models.Account, a models.Account) {
	length := len(s)

	if length > 1 {
		middle := length / 2

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			parallelMergeSort(s[:middle], a)
		}()

		go func() {
			defer wg.Done()
			parallelMergeSort(s[middle:], a)
		}()

		wg.Wait()
		merge(s, middle, a)
	}
}
