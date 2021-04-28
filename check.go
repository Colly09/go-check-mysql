package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"test-mysql/src/dbs"
)

func Println(str interface{}) {
	fmt.Println(fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 37, 41, str, 0x1B))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("请输入要检查的数据库名")
		fmt.Println("命令格式 ./check ucenter")
		return
	}
	dbConfig, err := ioutil.ReadFile("db.config")
	if err != nil {
		fmt.Println("获取数据库连接配置失败，请检查db.config")
		return
	}
	fmt.Println("数据库连接配置为：", string(dbConfig))
	fmt.Println("检查的数据库名为：", string(os.Args[1]))
	dbName := os.Args[1]

	dbs.ConnMysql(string(dbConfig), dbName)

	file, err := os.Open(fmt.Sprintf("%s.sql", dbName))
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var tableName string
	var completeTable []string
	var cloums []string
	var cloums2 []string
	var modify []string
	var cloumsMap map[string]string
	var pk string
	var index []string
	var checkPK bool

	filename := fmt.Sprintf("%s.modify.sql", dbName)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	for scanner.Scan() {
		lineText := scanner.Text()

		if strings.Contains(lineText, "CREATE ") {
			// 重置
			cloumsMap = make(map[string]string)
			cloums = cloums[0:0]
			completeTable = completeTable[0:0]
			cloums2 = cloums2[0:0]
			modify = modify[0:0]
			index = index[0:0]
			pk = ""
			checkPK = false

			// 开头
			line := strings.Split(lineText, "`")
			completeTable = append(completeTable, lineText)
			tableName = line[1]
			fmt.Println(fmt.Sprintf("检查 %s", tableName))

			dbCloums := dbs.GetCreate(tableName)
			_Line := strings.Split(dbCloums, "\n")

			// 表里面的结构
			for _, item := range _Line {
				if strings.Contains(item, "PRIMARY ") {
					// 主键
					line := strings.Split(item, "`")
					pk = line[1]
				} else if strings.Contains(item, "KEY ") {
					// 索引
					index = append(index, item)
				} else if strings.Index(item, "`") < 5 && strings.Index(item, "`") > 0 {
					line := strings.Split(item, "`")
					cloumsMap[line[1]] = item
					cloums = append(cloums, line[1])
				}
			}
		} else if strings.Contains(lineText, "DROP ") || strings.Contains(lineText, "SET ") {
			continue
		} else if strings.Contains(lineText, ";") {
			completeTable = append(completeTable, lineText)

			// 检查索引和主键
			if !checkPK && pk != "" {
				// 文件中没有主键 库里有主键
				modify = append(modify, "DROP PRIMARY KEY")
			}

			if len(cloums) == 0 {
				// 表不存在
				Println(fmt.Sprintf("表[%s]不存在，需创建", tableName))
				for _, item := range completeTable {
					Println(item)
					f.WriteString("\n")
					f.WriteString(item)
					f.WriteString("\n")
				}

				tableName = ""
				continue

			} else if len(cloums) > len(cloums2) {
				// 检查字段是否相等
				Println(fmt.Sprintf("表[%s]字段比标准多，请检查", tableName))
				Println("多的字段有：")
				for _, item := range cloums {
					exist := false
					for _, item2 := range cloums2 {
						if item == item2 {
							exist = true
						}
					}
					if !exist {
						Println(item)
					}
				}
			} else if len(cloums) < len(cloums2) {
				Println(fmt.Sprintf("表[%s]需要新增字段", tableName))
			}

			if len(modify) > 0 {
				// 检查有修改的字段
				Println(fmt.Sprintf("表[%s]字段、索引变更", tableName))
				Println(fmt.Sprintf("ALTER TABLE `%s`", tableName))
				str := strings.Replace(strings.Replace((strings.Join(modify, ",\n")+";"), ",,", ",", -1), ",;", ";", -1)
				Println(str)

				f.WriteString("\n")
				f.WriteString(fmt.Sprintf("ALTER TABLE `%s`", tableName))
				f.WriteString("\n")
				f.WriteString(str)
				f.WriteString("\n")
			}
			tableName = ""
			
		} else if tableName != "" {
			completeTable = append(completeTable, lineText)
			if strings.Contains(lineText, "PRIMARY ") {
				checkPK = true
				line := strings.Split(lineText, "`")
				if pk != line[1] {
					if pk != "" {
						modify = append(modify, "DROP PRIMARY KEY")
					}
					modify = append(modify, fmt.Sprintf("ADD PRIMARY KEY (`%s`)", line[1]))
				}
				/**
				  ADD INDEX `idx_test` (`i_id`, `i_name`) ;
					DROP INDEX `test`;
				*/
			} else if strings.Contains(lineText, "KEY ") {
				// 索引
				line := strings.Split(lineText, "`")
				keyOne := line[1]
				checkIndex := false
				for _, _index := range index {
					line := strings.Split(_index, "`")
					keyTwo := line[1]
					if keyOne == keyTwo {
						checkIndex = true
						//
						if strings.Replace(strings.Replace(lineText, ",", "", -1), " USING BTREE", "", -1) == strings.Replace(strings.Replace(_index, ",", "", -1), " USING BTREE", "", -1) {
							continue
						} else {
							modify = append(modify, fmt.Sprintf("DROP INDEX `%s`", keyTwo))
							modify = append(modify, fmt.Sprintf("ADD INDEX %s", strings.Trim(strings.Replace(lineText, "KEY ", "", 1), " ")))
							fmt.Println(lineText, _index)
						}
					}
				}

				if !checkIndex {
					modify = append(modify, fmt.Sprintf("ADD INDEX %s", strings.Trim(strings.Replace(lineText, "KEY ", "", 1), " ")))
				}

			} else {
				// 字段
				line := strings.Split(lineText, "`")
				// fmt.Println(line[1])
				cloums2 = append(cloums2, line[1])
				has := false
				_modify := false
				for _, item := range cloums {
					if item == line[1] {
						has = true
						if strings.Replace(lineText, ",", "", -1) != strings.Replace(cloumsMap[item], ",", "", -1) {
							_modify = true
						}
					}
				}

				if !has {
					// Println(fmt.Sprintf("字段缺失 %s", line[1]))
					modify = append(modify, fmt.Sprintf("ADD COLUMN %s", strings.Trim(lineText, " ")))
				}
				if _modify {
					modify = append(modify, fmt.Sprintf("MODIFY COLUMN %s", strings.Trim(lineText, " ")))
				}
			}

		}
	}
}
