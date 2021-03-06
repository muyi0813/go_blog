package models
import (
	"time"
	"github.com/astaxie/beego/orm"
	"strings"
	"blog/utils"
	"fmt"
	"strconv"
	"github.com/go-errors/errors"
)

type Article struct{
	Id int
	Title string
	Uri string
	Keywords string
	Abstract string
	Content string
	Author string
	Time time.Time
	Count int
	Status int
}

func (this *Article)TableName()string{
	return "article"
}
func init(){
	orm.RegisterModel(new(Article))
}
func AddArticle(title, content, keyword, abstract, author string)(int64, error){
	o := orm.NewOrm()

	sql := "insert into article(title, uri, keywords, abstract, content, author)values(?,?,?,?,?,?)"
	res, err := o.Raw(sql, title, strings.Replace(title, "/", "-", -1), keyword, abstract, content, author).Exec()
	if err != nil{
		return 0, err
	}else{
		return res.LastInsertId()
	}
}

func GetArticleById(id int)(Article, error){
	var art Article
	err := utils.GetCache("GetArticle.id."+fmt.Sprintf("%d", id), &art)
	if err != nil {
		o := orm.NewOrm()

		art = Article{Id: id}
		err = o.Read(&art, "id")
		utils.SetCache("GetArticle.id."+fmt.Sprintf("%d", id), &art, 600)
	}
	return art, err
}
func GetArticleByUri(uri string)(Article, error){
	var art Article
	err := utils.GetCache("GetArticleByUri.uri."+uri, &art)
	if err == nil {
		count, err := GetArticleViewCount(art.Id)
		if err == nil {
			art.Count = int(count)
		}
		return art, nil
	}else{
		o := orm.NewOrm()

		art = Article{Uri:uri}
		err = o.Read(&art, "uri")
		utils.SetCache("GetArticleByUri.uri."+uri, &art, 600)
	}
	return art, err
}
func GetArticleByTitle(title string)(Article, error){
	var art Article
	err := utils.GetCache("GetArticleByTitle.title."+title, &art)
	if err != nil{
		count, err := GetArticleViewCount(art.Id)
		if err == nil {
			art.Count = count
		}
		return art, nil
	}else{
		o := orm.NewOrm()

		art = Article{Title:title}
		err = o.Read(&art, "title")
		utils.SetCache("GetArticleByTitle.title"+title, art, 600)
	}
	return art, err
}

//获取浏览量
func GetArticleViewCount(id int)(int, error){
	var maps []orm.Params

	sql := `select count from article where id = ?`
	o := orm.NewOrm()
	num, err := o.Raw(sql, id).Values(&maps)
	if err == nil && num > 0 {
		count := maps[0]["count"].(string)
		return strconv.Atoi(count)
	}else{
		return 0, err
	}
}
func UpdateCount(id int)error{
	o := orm.NewOrm()

	art := Article{Id:id}
	err := o.Read(&art)
	o.QueryTable("article").Filter("id", id).Update(orm.Params{"count": art.Count+1})
	return err
}
func UpdateArticle(id int, uri string, newArt Article)error{
	if id == 0 && uri == "" {
		return errors.New("参数错误")
	}
	o := orm.NewOrm()

	var art Article
	if id != 0{
		art = Article{Id: id}
	}else{
		art = Article{Uri: uri}
	}
	art.Title = newArt.Title
	art.Keywords = newArt.Keywords
	art.Abstract = newArt.Abstract
	art.Content = newArt.Content

	getArt, _ := GetArticleById(int(id))
	utils.DelCache("GetArticleByUri.uri."+getArt.Uri)
	utils.DelCache("GetArticleByTitle.title."+getArt.Uri)
	utils.DelCache("GetArticleById.id." + strconv.Itoa(getArt.Id))

	_, err := o.Update(&art, "title", "keywords", "abstract", "content")
	return err
}
func DeleteArticle(id int64, uri string)(int64, error){
	if id == 0 && uri == "" {
		return 0, errors.New("参数错误")
	}
	o := orm.NewOrm()

	var art Article
	if id != 0{
		art.Id = int(id)
	}else{
		art.Uri = uri
	}
	getArt, _ := GetArticleById(int(id))
	utils.DelCache("GetArticleByUri.uri."+getArt.Uri)
	utils.DelCache("GetArticleByTitle.title."+getArt.Uri)
	utils.DelCache("GetArticleById.id." + strconv.Itoa(getArt.Id))

	return o.Delete(&art)
}
func CountByMonth()([]orm.Params, error){
	var maps []orm.Params
	err := utils.GetCache("CountByMonth", &maps)
	if err != nil {
		sql := `SELECT DATE_FORMAT(time, '%Y-%m') as date, count(*) as number, YEAR(time) as year, MONTH(time) as month
		FROM article GROUP BY date ORDER BY year DESC, month DESC`
		o := orm.NewOrm()

		num, err := o.Raw(sql).Values(&maps)
		if err == nil && num > 0{
			utils.SetCache("CountByMonth", maps, 3600)
			return maps, nil
		}else{
			return nil, err
		}
	}else{
		return maps, nil
	}
}