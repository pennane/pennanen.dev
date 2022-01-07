let usableCharacters = '';
let knownCharacters = '';

function fetchWords(path) {
    return new Promise(async (resolve, reject) => {
        let sanat = await fetch('kotus_sanat.txt')
        sanat = await sanat.text()
        sanat = sanat.split("\n")
        resolve(sanat)
    })
}

function findWordsWithInstructions(wordsArray, instructions) {
    let usableCharacters = instructions.usableCharacters.split('');
    let usableCharactersAmounts = new Map()

    usableCharacters.forEach(character => {
        if (usableCharactersAmounts.has(character)) usableCharactersAmounts.set(character, usableCharactersAmounts.get(character) + 1)
        else usableCharactersAmounts.set(character, 1)
    })

    let knownCharacters = instructions.knownCharacters.split('').map(character => ((character === '-' || character === '_' || character === ' ') ? null : character))

    let possibleWords = wordsArray

    //console.log('Initial amount of words', possibleWords.length)

    possibleWords = possibleWords.filter(word => word.length === knownCharacters.length)

    //console.log('Filtered words with wrong length', possibleWords.length)

    possibleWords = possibleWords.filter(word => {
        let characters = word.split('')

        const hasRightCharacters = character => usableCharacters.indexOf(character) >= 0;

        return characters.every(hasRightCharacters);
    })

    //console.log('Filtered words with incorrect letters', possibleWords.length)

    possibleWords = possibleWords.filter(word => {
        let characters = word.split('')
        let charactersAmounts = new Map()
        characters.forEach(character => {
            if (charactersAmounts.has(character)) charactersAmounts.set(character, charactersAmounts.get(character) + 1)
            else charactersAmounts.set(character, 1)
        })
        let passed = true;
        charactersAmounts.forEach((amount, character) => {
            if (usableCharactersAmounts.get(character) < amount) passed = false;
        })
        return passed;

    })
    //console.log('Filtered words that have too many of the correct letters', possibleWords.length)

    possibleWords = possibleWords.filter(word => {
        return knownCharacters.every((character, index) => character === null || word[index] === character)
    })

    //console.log('Possible words', possibleWords)
    return possibleWords
}

fetchWords('kotus_sanat.txt').then(words => {

    let foundWordsOutput = document.getElementById('found-words')

    function updateResults() {
        let instructions = knownCharacters.split(',').map(word => word.trim());
        let foundWords = [];

        instructions.forEach(instruction => {
            let newWords = findWordsWithInstructions(words, {
                usableCharacters,
                knownCharacters: instruction
            })
            foundWords = foundWords.concat(newWords)
        })

        foundWordsOutput.textContent = foundWords.join(', ')
    }

    let timeout_usableCharacters;
    let timeout_knownCharacters;

    let usableCharactersInput = document.getElementById('usable-characters');
    let knownCharactersInput = document.getElementById('known-characters')

    usableCharactersInput.addEventListener("input", () => {
        clearTimeout(timeout_usableCharacters);
        timeout_usableCharacters = setTimeout(() => {
            usableCharacters = usableCharactersInput.value.toLowerCase()
            updateResults()
        }, 600);
    })
    knownCharactersInput.addEventListener("input", () => {
        clearTimeout(timeout_knownCharacters);
        timeout_knownCharacters = setTimeout(() => {
            knownCharacters = knownCharactersInput.value.toLowerCase()
            updateResults()
        }, 600);
    })

    usableCharacters = usableCharactersInput.value.toLowerCase()
    knownCharacters = knownCharactersInput.value.toLowerCase()
    updateResults()

})

