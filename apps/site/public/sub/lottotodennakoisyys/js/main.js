let statistics = {
  v0: {
    payout: 0,
    amount: 0,
    probability: 0
  },
  v1: {
    payout: 0,
    amount: 0,
    probability: 0
  },
  v2: {
    payout: 0,
    amount: 0,
    probability: 0
  },
  v3: {
    payout: 0,
    amount: 0,
    probability: 0
  },
  v3b: {
    payout: 2,
    amount: 0,
    probability: 0
  },
  v4: {
    payout: 10,
    amount: 0,
    probability: 0
  },
  v5: {
    payout: 50,
    amount: 0,
    probability: 0
  },
  v6: {
    payout: 2000,
    amount: 0,
    probability: 0
  },
  v6b: {
    payout: 100000,
    amount: 0,
    probability: 0
  },
  v7: {
    payout: 5000000,
    amount: 0,
    probability: 0
  },
  total: {
    bonus: 0,
    amount: 0,
    wins: 0,
    probability: 0
  }
};

let lotteryNumberElements = [
  document.getElementById("lspan1"),
  document.getElementById("lspan2"),
  document.getElementById("lspan3"),
  document.getElementById("lspan4"),
  document.getElementById("lspan5"),
  document.getElementById("lspan6"),
  document.getElementById("lspan7")
];

let statisticsElements = {
  "3+1": {
    amount: document.getElementById("v3ba"),
    probability: document.getElementById("v3bp"),
    winnings: document.getElementById("v3bw")
  },
  "4": {
    amount: document.getElementById("v4a"),
    probability: document.getElementById("v4p"),
    winnings: document.getElementById("v4w")
  },
  "5": {
    amount: document.getElementById("v5a"),
    probability: document.getElementById("v5p"),
    winnings: document.getElementById("v5w")
  },
  "6": {
    amount: document.getElementById("v6a"),
    probability: document.getElementById("v6p"),
    winnings: document.getElementById("v6w")
  },
  "6+1": {
    amount: document.getElementById("v6ba"),
    probability: document.getElementById("v6bp"),
    winnings: document.getElementById("v6bw")
  },
  "7": {
    amount: document.getElementById("v7a"),
    probability: document.getElementById("v7p"),
    winnings: document.getElementById("v7w")
  },
  total: {
    rows: document.getElementById("yhta"),
    wins: document.getElementById("voita"),
    probability: document.getElementById("mahdp"),
    used: document.getElementById("used"),
    gained: document.getElementById("gained")
  }
};

const displayWinningNumbers = numbers => {
  let bonus = numbers[7];
  let nums = numbers.slice(0, 7).sort((a, b) => a - b);
  console.log(`Lottonumerot: ${nums}`);
  console.log(`Lisänumero: ${bonus}`);
};

const groupInt = int => {
  if (int === Infinity || int === 0 || int < 999) return int;
  int = parseInt(int);
  let skip = ["+", "e", ".", ","];
  let strInt = int.toString();
  let groupIndex = 1;
  let toReturn = "";
  for (let i = strInt.length; i > 0; i--) {
    toReturn = strInt[i - 1] + toReturn + "";
    if (skip.indexOf(strInt[i - 1]) > -1) {
      groupIndex = 1;
      continue;
    }
    if (groupIndex % 3 === 0) {
      toReturn = " " + toReturn;
    }
    groupIndex++;
  }
  return toReturn;
};

const generateNumbers = (amount = 7, highest = 40, lowest = 1) => {
  let numbers = [];
  do {
    let n = Math.floor(Math.random() * highest) + lowest;
    if (numbers.indexOf(n) === -1) {
      numbers.push(n);
    }
  } while (numbers.length < amount);

  return numbers;
};

const getHowManyCorrect = numbers => {
  let correctNumbers = 0;
  let bonus = false;
  for (let i = 0; i < 7; i++) {
    if (numbers.indexOf(winningNumbers[i]) !== -1) correctNumbers++;
  }
  if (numbers.indexOf(winningNumbers[7]) !== -1) bonus = true;
  return {
    correct: correctNumbers,
    bonus: bonus
  };
};

const purchaseTickets = (amount = 1) => {
  for (let i = 0; i < amount; i++) {
    let numbers = generateNumbers();
    let correctNumbers = getHowManyCorrect(numbers);
    updateStatistics(correctNumbers);
    if (i === amount - 1) {
      purchasedRow = numbers.sort((a, b) => a - b);
      updateDom();
    }
  }
};

const purchaseInfinitely = () => {
  let domTarget = document.getElementById("purchaseinfiniteticket");
  if (purchasingInfinitely === false) {
    purchasingInfinitely = true;

    domTarget.textContent = "Pysäytä rivien osto";
    domTarget.classList.toggle("active", true);

    let w7n = statistics.v7.amount;

    window.lotteryloop = setInterval(() => {
      purchaseTickets(1000);
      if (w7n !== statistics.v7.amount) {
        clearInterval(lotteryloop);
      }
    }, 40);
  } else if (purchasingInfinitely === true) {
    purchasingInfinitely = false;

    domTarget.textContent = "Osta loputtomasti rivejä";
    domTarget.classList.toggle("active", false);

    clearInterval(lotteryloop);
  }
};

const updateStatistics = ({ correct, bonus }) => {
  statistics.total.amount++;

  if (bonus === true) {
    statistics.total.bonus++;
  }

  let stat =
    bonus && (correct === 3 || correct === 6) ? `v${correct}b` : `v${correct}`;

  statistics[stat].amount++;

  if ((correct === 3 && bonus) || correct > 3) {
    statistics.total.wins++;
  }

  statistics.total.probability = Number(
    (1 / (statistics.total.wins / statistics.total.amount)).toPrecision(3)
  );
  statistics.v3b.probability = Number(
    (1 / (statistics.v3b.amount / statistics.total.amount)).toPrecision(4)
  );
  statistics.v4.probability = Number(
    (1 / (statistics.v4.amount / statistics.total.amount)).toPrecision(4)
  );
  statistics.v5.probability = Number(
    (1 / (statistics.v5.amount / statistics.total.amount)).toPrecision(4)
  );
  statistics.v6.probability = Number(
    (1 / (statistics.v6.amount / statistics.total.amount)).toPrecision(6)
  );
  statistics.v6b.probability = Number(
    (1 / (statistics.v6b.amount / statistics.total.amount)).toPrecision(9)
  );
  statistics.v7.probability = Number(
    (1 / (statistics.v7.amount / statistics.total.amount)).toPrecision(10)
  );
};

const updateDom = () => {
  lotteryNumberElements.forEach((el, i) => {
    el.textContent = purchasedRow[i];
  });

  statisticsElements["3+1"].amount.textContent = groupInt(
    statistics.v3b.amount
  );
  statisticsElements["3+1"].probability.textContent =
    statistics.v3b.probability;
  statisticsElements["3+1"].winnings.textContent = groupInt(
    statistics.v3b.amount * statistics.v3b.payout
  );

  statisticsElements["4"].amount.textContent = groupInt(statistics.v4.amount);
  statisticsElements["4"].probability.textContent = statistics.v4.probability;
  statisticsElements["4"].winnings.textContent = groupInt(
    statistics.v4.amount * statistics.v4.payout
  );

  statisticsElements["5"].amount.textContent = groupInt(statistics.v5.amount);
  statisticsElements["5"].probability.textContent = groupInt(
    statistics.v5.probability
  );
  statisticsElements["5"].winnings.textContent = groupInt(
    statistics.v5.amount * statistics.v5.payout
  );

  statisticsElements["6"].amount.textContent = groupInt(statistics.v6.amount);
  statisticsElements["6"].probability.textContent = groupInt(
    statistics.v6.probability
  );
  statisticsElements["6"].winnings.textContent = groupInt(
    statistics.v6.amount * statistics.v6.payout
  );

  statisticsElements["6+1"].amount.textContent = groupInt(
    statistics.v6b.amount
  );
  statisticsElements["6+1"].probability.textContent = groupInt(
    statistics.v6b.probability
  );
  statisticsElements["6+1"].winnings.textContent = groupInt(
    statistics.v6b.amount * statistics.v6b.payout
  );

  statisticsElements["7"].amount.textContent = groupInt(statistics.v7.amount);
  statisticsElements["7"].probability.textContent = groupInt(
    statistics.v7.probability
  );
  statisticsElements["7"].winnings.textContent = groupInt(
    statistics.v7.amount * statistics.v7.payout
  );

  statisticsElements["total"].rows.textContent = groupInt(
    statistics.total.amount
  );
  statisticsElements["total"].wins.textContent = groupInt(
    statistics.total.wins
  );
  statisticsElements["total"].probability.textContent =
    statistics.total.probability;
  statisticsElements["total"].used.textContent = groupInt(
    statistics.total.amount
  );
  statisticsElements["total"].gained.textContent = groupInt(
    statistics.v7.amount * statistics.v7.payout +
      statistics.v6b.amount * statistics.v6b.payout +
      statistics.v6.amount * statistics.v6.payout +
      statistics.v5.amount * statistics.v5.payout +
      statistics.v4.amount * statistics.v4.payout +
      statistics.v3b.amount * statistics.v3b.payout
  );
};

let purchasedRow = [];
let purchasingInfinitely = false;
let winningNumbers = generateNumbers(8);

displayWinningNumbers(winningNumbers);

[...document.querySelectorAll(".fraction")].forEach(el => {
  let split = el.innerHTML.split("/");
  if (split.length === 2) {
    el.innerHTML =
      '<span class="top">' +
      split[0] +
      '</span><span class="bottom">' +
      split[1] +
      "</span>";
  }
});

document.getElementById("purchase1ticket").addEventListener("click", () => {
  purchaseTickets(1);
});
document.getElementById("purchase10ticket").addEventListener("click", () => {
  purchaseTickets(10);
});
document.getElementById("purchase100ticket").addEventListener("click", () => {
  purchaseTickets(100);
});
document.getElementById("purchase1000ticket").addEventListener("click", () => {
  purchaseTickets(1000);
});
document
  .getElementById("purchaseinfiniteticket")
  .addEventListener("click", () => {
    purchaseInfinitely();
  });
