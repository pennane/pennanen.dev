function _toConsumableArray(arr) {
  return (
    _arrayWithoutHoles(arr) ||
    _iterableToArray(arr) ||
    _unsupportedIterableToArray(arr) ||
    _nonIterableSpread()
  );
}

function _nonIterableSpread() {
  throw new TypeError(
    "Invalid attempt to spread non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."
  );
}

function _unsupportedIterableToArray(o, minLen) {
  if (!o) return;
  if (typeof o === "string") return _arrayLikeToArray(o, minLen);
  var n = Object.prototype.toString.call(o).slice(8, -1);
  if (n === "Object" && o.constructor) n = o.constructor.name;
  if (n === "Map" || n === "Set") return Array.from(o);
  if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))
    return _arrayLikeToArray(o, minLen);
}

function _iterableToArray(iter) {
  if (typeof Symbol !== "undefined" && Symbol.iterator in Object(iter))
    return Array.from(iter);
}

function _arrayWithoutHoles(arr) {
  if (Array.isArray(arr)) return _arrayLikeToArray(arr);
}

function _arrayLikeToArray(arr, len) {
  if (len == null || len > arr.length) len = arr.length;
  for (var i = 0, arr2 = new Array(len); i < len; i++) {
    arr2[i] = arr[i];
  }
  return arr2;
}

var statistics = {
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
var lotteryNumberElements = [
  document.getElementById("lspan1"),
  document.getElementById("lspan2"),
  document.getElementById("lspan3"),
  document.getElementById("lspan4"),
  document.getElementById("lspan5"),
  document.getElementById("lspan6"),
  document.getElementById("lspan7")
];
var statisticsElements = {
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

var displayWinningNumbers = function displayWinningNumbers(numbers) {
  var bonus = numbers[7];
  var nums = numbers.slice(0, 7).sort(function(a, b) {
    return a - b;
  });
  console.log("Lottonumerot: ".concat(nums));
  console.log("Lis\xE4numero: ".concat(bonus));
};

var groupInt = function groupInt(_int) {
  if (_int === Infinity || _int === 0 || _int < 999) return _int;
  _int = parseInt(_int);
  var skip = ["+", "e", ".", ","];

  var strInt = _int.toString();

  var groupIndex = 1;
  var toReturn = "";

  for (var i = strInt.length; i > 0; i--) {
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

var generateNumbers = function generateNumbers() {
  var amount =
    arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : 7;
  var highest =
    arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : 40;
  var lowest =
    arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : 1;
  var numbers = [];

  do {
    var n = Math.floor(Math.random() * highest) + lowest;

    if (numbers.indexOf(n) === -1) {
      numbers.push(n);
    }
  } while (numbers.length < amount);

  return numbers;
};

var getHowManyCorrect = function getHowManyCorrect(numbers) {
  var correctNumbers = 0;
  var bonus = false;

  for (var i = 0; i < 7; i++) {
    if (numbers.indexOf(winningNumbers[i]) !== -1) correctNumbers++;
  }

  if (numbers.indexOf(winningNumbers[7]) !== -1) bonus = true;
  return {
    correct: correctNumbers,
    bonus: bonus
  };
};

var purchaseTickets = function purchaseTickets() {
  var amount =
    arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : 1;

  for (var i = 0; i < amount; i++) {
    var numbers = generateNumbers();
    var correctNumbers = getHowManyCorrect(numbers);
    updateStatistics(correctNumbers);

    if (i === amount - 1) {
      purchasedRow = numbers.sort(function(a, b) {
        return a - b;
      });
      updateDom();
    }
  }
};

var purchaseInfinitely = function purchaseInfinitely() {
  var domTarget = document.getElementById("purchaseinfiniteticket");

  if (purchasingInfinitely === false) {
    purchasingInfinitely = true;
    domTarget.textContent = "Pysäytä rivien osto";
    domTarget.classList.toggle("active", true);
    var w7n = statistics.v7.amount;
    window.lotteryloop = setInterval(function() {
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

var updateStatistics = function updateStatistics(_ref) {
  var correct = _ref.correct,
    bonus = _ref.bonus;
  statistics.total.amount++;

  if (bonus === true) {
    statistics.total.bonus++;
  }

  var stat =
    bonus && (correct === 3 || correct === 6)
      ? "v".concat(correct, "b")
      : "v".concat(correct);
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

var updateDom = function updateDom() {
  lotteryNumberElements.forEach(function(el, i) {
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

var purchasedRow = [];
var purchasingInfinitely = false;
var winningNumbers = generateNumbers(8);
displayWinningNumbers(winningNumbers);

_toConsumableArray(document.querySelectorAll(".fraction")).forEach(function(
  el
) {
  var split = el.innerHTML.split("/");

  if (split.length === 2) {
    el.innerHTML =
      '<span class="top">' +
      split[0] +
      '</span><span class="bottom">' +
      split[1] +
      "</span>";
  }
});

document
  .getElementById("purchase1ticket")
  .addEventListener("click", function() {
    purchaseTickets(1);
  });
document
  .getElementById("purchase10ticket")
  .addEventListener("click", function() {
    purchaseTickets(10);
  });
document
  .getElementById("purchase100ticket")
  .addEventListener("click", function() {
    purchaseTickets(100);
  });
document
  .getElementById("purchase1000ticket")
  .addEventListener("click", function() {
    purchaseTickets(1000);
  });
document
  .getElementById("purchaseinfiniteticket")
  .addEventListener("click", function() {
    purchaseInfinitely();
  });
