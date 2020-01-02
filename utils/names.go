package utils

var (
	//姓氏
	familyNames = []string{"赵", "钱", "孙", "李", "周", "吴", "郑", "王", "冯", "陈", "楚", "卫", "蒋", "沈", "韩", "杨", "张", "易"}
	//辈分
	middleNamesMap = map[string][]string{}
	//名字
	lastNames = []string{"春", "夏", "秋", "冬", "风", "霜", "雨", "雪", "木", "禾", "米", "竹", "山", "石", "田", "土", "福", "禄", "寿", "喜", "文", "武", "才", "华"}
)

func init() {
	middleNamesMap["赵"] = []string{"大", "国", "益", "之", "仕", "世", "秉", "忠", "德", "全", "立", "志", "承", "先", "泽", "诗", "书", "继", "祖", "传", "代", "远", "永", "佑", "启", "家", "邦", "振", "万", "年"}
	middleNamesMap["欧阳"] = []string{"元", "梦", "应", "祖", "子", "添", "永", "秀", "文", "才", "思", "颜", "承", "德", "正", "道", "积", "享", "荣", "华", "洪", "范", "征", "恩", "锡", "彝", "伦", "叙", "典", "常"}
	for _, x := range familyNames {
		if x != "欧阳" {
			middleNamesMap[x] = middleNamesMap["赵"]
		} else {
			middleNamesMap[x] = middleNamesMap["欧阳"]
		}
	}
}

//GetRandomName 获取随机名字
func GetRandomName() (name string) {
	familyName := familyNames[GetRandomInt(0, len(familyNames)-1)]
	middleName := middleNamesMap[familyName][GetRandomInt(0, len(middleNamesMap[familyName])-1)]
	lastName := lastNames[GetRandomInt(0, len(lastNames)-1)]
	return familyName + middleName + lastName
}
