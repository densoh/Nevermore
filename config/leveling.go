package config

import "math"

var LevelCap = 25
var SkillCap = 36000000

// TierExpLevels Leveling Values
var TierExpLevels = map[int]int{
	2:  450,       //450
	3:  1500,      //1050
	4:  4000,      //2500
	5:  10000,     //6000
	6:  25000,     //15000
	7:  75000,     //50000
	8:  225000,    //150000
	9:  500000,    //275000
	10: 1000000,   //500000
	11: 1800000,   //800000
	12: 3000000,   //1200000
	13: 4800000,   //1800000
	14: 7400000,   //2600000
	15: 11000000,  //3600000
	16: 15800000,  //4800000
	17: 21800000,  //6000000
	18: 29000000,  //7200000
	19: 37400000,  //8400000
	20: 47400000,  //10000000
	21: 59800000,  //12400000
	22: 75000000,  //15200000
	23: 93000000,  //18000000
	24: 117000000, //24000000
	25: 147000000, //30000000
}

// MobExpByLevel is the exp a mob of each level is worth at an exp_multiplier
// of 1. The values are set so the solo kills a tier N player needs at level N
// mobs (comment, at the 0.955 average solo award) rise smoothly and are
// higher every tier than the one below. They depend on the TierExpLevels
// gaps, so retune them together.
var MobExpByLevel = map[int]int{
	1:  10,    //47 kills
	2:  23,    //48 kills
	3:  49,    //53 kills
	4:  92,    //68 kills
	5:  155,   //101 kills
	6:  285,   //184 kills
	7:  550,   //286 kills
	8:  710,   //406 kills
	9:  920,   //569 kills
	10: 1050,  //798 kills
	11: 1150,  //1093 kills
	12: 1350,  //1396 kills
	13: 1650,  //1650 kills
	14: 2125,  //1774 kills
	15: 2725,  //1845 kills
	16: 3350,  //1875 kills
	17: 4000,  //1885 kills
	18: 4600,  //1912 kills
	19: 5300,  //1976 kills
	20: 6075,  //2137 kills
	21: 6725,  //2367 kills
	22: 7300,  //2582 kills
	23: 9450,  //2659 kills
	24: 11700, //2685 kills
	25: 13000,
}

// MobExpGrowthPastCap is the per-level growth of MobExpByLevel past level 25.
var MobExpGrowthPastCap = 1.10

// MobBaseExperience is MobExpByLevel for level, treating levels under 1 as 1
// and growing by MobExpGrowthPastCap per level past the end of the table.
func MobBaseExperience(level int) int {
	if level < 1 {
		level = 1
	}
	if level <= 25 {
		return MobExpByLevel[level]
	}
	return int(math.Round(float64(MobExpByLevel[25]) * math.Pow(MobExpGrowthPastCap, float64(level-25))))
}

func MaxLoss(tier int) int {
	return TierExpLevels[tier+1] - TierExpLevels[tier]
}

/* Old Values
// TierExpLevels Leveling Values
var TierExpLevels = map[int]int{
	2:  650,
	3:  2800,
	4:  7500,
	5:  15000,
	6:  30000,
	7:  84000,
	8:  250000,
	9:  550000,
	10: 1100000,
	11: 1900000,
	12: 3000000,
	13: 4500000,
	14: 6500000,
	15: 8600000,
	16: 11200000,
	17: 14650000,
	18: 19400000,
	19: 25300000,
	20: 32500000,
	21: 41100000,
	22: 51200000,
	23: 63100000,
	24: 76900000,
	25: 92500000,
}
*/

var GoldPerLevel = map[int]int{
	2:  100,
	3:  333,
	4:  666,
	5:  1333,
	6:  2666,
	7:  5333,
	8:  10666,
	9:  21333,
	10: 42666,
	11: 84333,
	12: 142666,
	13: 210333,
	14: 250666,
	15: 275333,
	16: 300666,
	17: 345333,
	18: 480666,
	19: 596333,
	20: 723666,
	21: 876333,
	22: 1010666,
	23: 1212333,
	24: 1433666,
	25: 1566333,
}

var TextTiers = []string{
	"zero",
	"first",
	"second",
	"third",
	"fourth",
	"fifth",
	"sixth",
	"seventh",
	"eighth",
	"ninth",
	"tenth",
	"eleventh",
	"twelfth",
	"thirteenth",
	"fourteenth",
	"fifteenth",
	"sixteenth",
	"seventeenth",
	"eighteenth",
	"nineteenth",
	"twentieth",
	"twenty-first",
	"twenty-second",
	"twenty-third",
	"twenty-fourth",
	"twenty-fifth",
	"twenty-sixth",
	"twenty-seventh",
	"twenty-eighth",
	"twenty-ninth",
	"thirtieth",
	"thirty-first",
	"thirty-second",
	"thirty-third",
	"thirty-fourth",
	"thirty-fifth",
	"thirty-sixth",
	"thirty-seventh",
	"thirty-eighth",
	"thirty-ninth",
	"fortieth",
	"forty-first",
	"forty-second",
	"forty-third",
	"forty-fourth",
	"forty-fifth",
	"forty-sixth",
	"forty-seventh",
	"forty-eighth",
	"forty-ninth",
	"fiftieth",
}

var PrintNumbers = []string{
	"0",
	"1st",
	"2nd",
	"3rd",
	"4th",
	"5th",
	"6th",
	"7th",
	"8th",
	"9th",
	"10th",
	"11th",
	"12th",
	"13th",
	"14th",
	"15th",
	"16th",
	"17th",
	"18th",
	"19th",
	"20th",
	"21st",
	"22nd",
	"23rd",
	"24th",
	"25th",
	"26th",
	"27th",
	"28th",
	"29th",
	"30th",
	"31st",
	"32nd",
	"33rd",
	"34th",
	"35th",
	"36th",
	"37th",
	"38th",
	"39th",
	"40th",
}

var TextNumbers = []string{
	"zero",
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
	"ten",
	"eleven",
	"twelve",
	"thirteen",
	"fourteen",
	"fifteen",
	"sixteen",
	"seventeen",
	"eighteen",
	"nineteen",
	"twenty",
	"twenty-one",
	"twenty-two",
	"twenty-three",
	"twenty-four",
	"twenty-five",
	"twenty-six",
	"twenty-seven",
	"twenty-eight",
	"twenty-nine",
	"thirty",
	"thirty-one",
	"thirty-two",
	"thirty-three",
	"thirty-four",
	"thirty-five",
	"thirty-six",
	"thirty-seven",
	"thirty-eight",
	"thirty-nine",
	"forty",
	"forty-one",
	"forty-two",
	"forty-three",
	"forty-four",
	"forty-five",
	"forty-six",
	"forty-seven",
	"forty-eight",
	"forty-nine",
	"fifty",
}

var HealingHandCost = map[int]int{
	1:  0,
	2:  20,
	3:  67,
	4:  133,
	5:  267,
	6:  533,
	7:  1067,
	8:  2133,
	9:  4267,
	10: 8533,
	11: 16867,
	12: 28533,
	13: 42067,
	14: 50133,
	15: 55067,
	16: 60133,
	17: 69067,
	18: 96133,
	19: 19267,
	20: 144733,
	21: 175267,
	22: 202133,
	23: 242467,
	24: 286733,
	25: 313267,
}
