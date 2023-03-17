package main

import (
	"Personal-web/connection"
	"Personal-web/middleware"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Template struct {
	templates *template.Template
}

type Projects struct {
	Id          int
	Title       string
	Sdate       time.Time
	Edate       time.Time
	SDconvert   string
	EDconvert   string
	Duration    string
	Descript    string
	Technologys []string
	Image       string
	Author      int
	Owner       interface{}
}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	connection.ConnectDB()
	e := echo.New()

	// root statis untuk mengakses folder public
	e.Static("/public", "public")   //public
	e.Static("/uploads", "uploads") //upload image

	// untuk menggunakan echo session
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("session"))))

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}

	// renderer
	e.Renderer = t

	// routing
	e.GET("/", home)
	e.GET("/contact", contactMe)
	e.GET("/project", myProject)
	e.GET("/project-detail/:id", projectDetail) //:id => url params
	e.POST("/add-project", middleware.UploadFile(addProject))
	e.GET("/delete/:id", deleteProject)              //:id => url params
	e.GET("/edit-project/:id", editProject)          //:id => url params
	e.POST("/edit/:id", middleware.UploadFile(edit)) //:id => url params
	e.GET("/form-register", formRegister)
	e.GET("/form-login", formLogin)
	e.POST("/register", addRegister)
	e.POST("/login", login)
	e.GET("/logout", logout)

	fmt.Println("localhost: 5000 sucssesfully")
	e.Logger.Fatal(e.Start("localhost: 5000"))
}

func formRegister(c echo.Context) error {
	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] == true {
		redirectMessage(c, "You need to log out first to register new data !", false, "/")
	}

	flash := map[string]interface{}{
		"FlashStatus":  sess.Values["status"],
		"FlashMessage": sess.Values["message"],
		"IsLoggins":    sess.Values["isLogin"],
		"FlashName":    sess.Values["name"],
	}

	delete(sess.Values, "message")
	sess.Save(c.Request(), c.Response())

	return c.Render(http.StatusOK, "formRegister.html", flash)
}

func addRegister(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	var emailValidate string
	err := connection.Conn.QueryRow(context.Background(), "SELECT email from tb_user WHERE email=$1", email).Scan(&emailValidate)

	// check if there is already email in database
	if err != nil {
		// register user and store user data to database
		_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user (name, email, pasword) VALUES ($1,$2,$3)", name, email, passwordHash)

		if err != nil {
			redirectMessage(c, "Register Failed !", false, "/form-register")
		}
		// if there is no error encounter, the code will go through
		redirectMessage(c, "Register Success !", true, "/form-login")

	} else {
		// if there is already email in database
		redirectMessage(c, "Email Is taken", false, "/form-register")
	}

	return c.Redirect(http.StatusMovedPermanently, "/form-register")
}

func home(c echo.Context) error {
	sess, _ := session.Get("session", c)

	data, _ := connection.Conn.Query(context.Background(), "SELECT tb_project.id, tb_project.title, start_date, end_date, technologys, description, image, author_id FROM tb_project INNER JOIN tb_user ON tb_project.author_id = tb_user.id ORDER BY tb_project.id DESC")

	var result []Projects
	for data.Next() {
		var each = Projects{}

		err := data.Scan(&each.Id, &each.Title, &each.Sdate, &each.Edate, &each.Technologys, &each.Descript, &each.Image, &each.Author)

		if err != nil {
			fmt.Println(err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"Message ": err.Error()})
		}

		durasi := each.Edate.Sub(each.Sdate)
		var Durations string

		if durasi.Hours()/24 < 7 {
			Durations = strconv.FormatFloat(durasi.Hours()/24, 'f', 0, 64) + " Days"
		} else if durasi.Hours()/24/7 < 4 {
			Durations = strconv.FormatFloat(durasi.Hours()/24/7, 'f', 0, 64) + " Weeks"
		} else if durasi.Hours()/24/30 < 12 {
			Durations = strconv.FormatFloat(durasi.Hours()/24/30, 'f', 0, 64) + " Months"
		} else {
			Durations = strconv.FormatFloat(durasi.Hours()/24/30/12, 'f', 0, 64) + " Years"
		}

		each.Duration = Durations

		result = append(result, each)
	}
	project := map[string]interface{}{
		"Projects":     result,
		"IsLoggins":    sess.Values["isLogin"],
		"FlashMessage": sess.Values["message"],
		"FlashStatus":  sess.Values["status"],
		"FlashName":    sess.Values["name"],
		"FlashUserId":  sess.Values["id"],
	}
	delete(sess.Values, "message")
	sess.Save(c.Request(), c.Response())

	return c.Render(http.StatusOK, "index.html", project)
} //clear

func contactMe(c echo.Context) error {
	sess, _ := session.Get("session", c)

	send := map[string]interface{}{
		"IsLoggins": sess.Values["isLogin"],
		"FlashName": sess.Values["name"],
	}
	return c.Render(http.StatusOK, "contact-me.html", send)
}

func myProject(c echo.Context) error {
	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] != true {
		redirectMessage(c, "You need to login to access it!", false, "/login")
	}

	send := map[string]interface{}{
		"FlashMessage": sess.Values["message"],
		"FlashStatus":  sess.Values["status"],
		"IsLoggins":    sess.Values["isLogin"],
		"FlashName":    sess.Values["name"],
	}

	delete(sess.Values, "message")
	sess.Save(c.Request(), c.Response())

	return c.Render(http.StatusOK, "myProject.html", send)
} //clear

func addProject(c echo.Context) error {
	sess, _ := session.Get("session", c)

	title := c.FormValue("project-name")
	sDate := c.FormValue("start-date")
	eDate := c.FormValue("end-date")
	tech := c.Request().Form["check"]
	desc := c.FormValue("description")

	image := c.Get("dataFile").(string)
	author := sess.Values["id"]

	_, err := connection.Conn.Exec(context.Background(), "INSERT INTO tb_project (title, start_date, end_date, technologys, description, image, author_id) VALUES ($1, $2, $3, $4, $5, $6, $7)", title, sDate, eDate, tech, desc, image, author)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message ": err.Error()})
	}
	return c.Redirect(http.StatusMovedPermanently, "/")
} //clear

func projectDetail(c echo.Context) error {
	sess, _ := session.Get("session", c)
	id, _ := strconv.Atoi(c.Param("id"))

	var ProjectDetail = Projects{}

	err := connection.Conn.QueryRow(context.Background(), "SELECT id, title, start_date, end_date, technologys, description, image FROM tb_project WHERE id=$1", id).Scan(&ProjectDetail.Id, &ProjectDetail.Title, &ProjectDetail.Sdate, &ProjectDetail.Edate, &ProjectDetail.Technologys, &ProjectDetail.Descript, &ProjectDetail.Image)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message ": err.Error()})
	}

	// convert date ke string
	sDateFormat := ProjectDetail.Sdate.Format("02 January 2006")
	eDateFormat := ProjectDetail.Edate.Format("02 January 2006")

	// SET Durations
	durasi := ProjectDetail.Edate.Sub(ProjectDetail.Sdate)
	var Durations string

	if durasi.Hours()/24 < 7 {
		Durations = strconv.FormatFloat(durasi.Hours()/24, 'f', 0, 64) + " Days"
	} else if durasi.Hours()/24/7 < 4 {
		Durations = strconv.FormatFloat(durasi.Hours()/24/7, 'f', 0, 64) + " Weeks"
	} else if durasi.Hours()/24/30 < 12 {
		Durations = strconv.FormatFloat(durasi.Hours()/24/30, 'f', 0, 64) + " Months"
	} else {
		Durations = strconv.FormatFloat(durasi.Hours()/24/30/12, 'f', 0, 64) + " Years"
	}

	send := map[string]interface{}{
		"Projects":  ProjectDetail,
		"Duration":  Durations,
		"StartD":    sDateFormat,
		"EndD":      eDateFormat,
		"IsLoggins": sess.Values["isLogin"],
		"FlashName": sess.Values["name"],
	}
	return c.Render(http.StatusOK, "projectDetail.html", send)
} //clear

func deleteProject(c echo.Context) error {
	sess, _ := session.Get("session", c)
	id, _ := strconv.Atoi(c.Param("id"))

	var author int

	err := connection.Conn.QueryRow(context.Background(), "SELECT user_id FROM tb_projects WHERE id=$1", id).Scan(&author)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if sess.Values["id"] == author {
		_, err := connection.Conn.Exec(context.Background(), "DELETE FROM public.tb_projects WHERE id=$1", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"Message": err.Error()})
		}

		return c.Redirect(http.StatusMovedPermanently, "/")
	} else {
		redirectMessage(c, "You can't delete this data !", false, "/")
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
} //clear

func editProject(c echo.Context) error {
	sess, _ := session.Get("session", c)
	if sess.Values["isLogin"] != true {
		redirectMessage(c, "You need to login to access it!", false, "/form-login")
	}
	id, _ := strconv.Atoi(c.Param("id"))

	edit := Projects{}
	err := connection.Conn.QueryRow(context.Background(), "SELECT id, Title, start_date, end_date, technologys, description, image, author_id FROM tb_project WHERE id=$1;", id).Scan(&edit.Id, &edit.Title, &edit.Sdate, &edit.Edate, &edit.Technologys, &edit.Descript, &edit.Image, &edit.Author)

	if sess.Values["id"] != edit.Author {
		redirectMessage(c, "You can't access this data !", false, "/")
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message ": err.Error()})
	}

	var Python, Js, React, Node bool

	for _, techno := range edit.Technologys {
		if techno == "python" {
			Python = true
		}
		if techno == "js" {
			Js = true
		}
		if techno == "react" {
			React = true
		}
		if techno == "node" {
			Node = true
		}
	}
	StartFormat := edit.Sdate.Format("2006-01-02")
	EndFormat := edit.Edate.Format("2006-01-02")

	editResult := map[string]interface{}{
		"Edit":         edit,
		"Id":           id,
		"StartD":       StartFormat,
		"EndD":         EndFormat,
		"Tech1":        Python,
		"Tech2":        Js,
		"Tech3":        React,
		"Tech4":        Node,
		"IsLoggins":    sess.Values["isLogin"],
		"FlashName":    sess.Values["name"],
		"FlashMessage": sess.Values["message"],
		"FlashStatus":  sess.Values["status"],
	}

	delete(sess.Values, "message")
	sess.Save(c.Request(), c.Response())

	return c.Render(http.StatusOK, "updateProject.html", editResult)
} //clear

func edit(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	title := c.FormValue("project-name")
	SDate := c.FormValue("start-date")
	EDate := c.FormValue("end-date")
	descript := c.FormValue("description")
	technologys := c.Request().Form["check"]
	image := c.Get("dataFile").(string)

	sdate, _ := time.Parse("2006-01-02", SDate)
	edate, _ := time.Parse("2006-01-02", EDate)

	_, err := connection.Conn.Exec(context.Background(), "UPDATE public.tb_project SET title=$1, start_date=$2, end_date=$3, description=$4, technologys=$5, image=$6 WHERE id=$7;", title, sdate, edate, descript, technologys, image, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message ": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
} //clear

func formLogin(c echo.Context) error {
	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] == true {
		redirectMessage(c, "You need to log out first to login again !", false, "/")
	}

	flash := map[string]interface{}{
		"FlashStatus":  sess.Values["status"],
		"FlashMessage": sess.Values["message"],
		"IsLoggins":    sess.Values["isLogin"],
		"FlashName":    sess.Values["name"],
	}

	delete(sess.Values, "message")
	sess.Save(c.Request(), c.Response())

	return c.Render(http.StatusOK, "formLogin.html", flash)
}

func login(c echo.Context) error {

	email := c.FormValue("email")
	password := c.FormValue("password")

	user := User{}

	err := connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return redirectMessage(c, "Couldn't find Email", false, "/form-login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return redirectMessage(c, "Wrong Password!", false, "/form-login")
	}

	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = 10800 // umur sesi login yg akan bertahan, 10800 detik diconvert jadi 3 jam
	sess.Values["message"] = "Login Success !"
	sess.Values["status"] = true //show alert
	sess.Values["name"] = user.Name
	sess.Values["id"] = user.ID
	sess.Values["isLogin"] = true //akses login
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func redirectMessage(c echo.Context, message string, status bool, path string, id ...string) error {

	// adding optional parameters
	// if id is not provided use an empty string
	var idValue string
	if len(id) > 0 {
		idValue = id[0]
	}

	// if id is provided, replace $1 to idValue
	if idValue != "" {
		path = strings.Replace(path, "$1", idValue, 1)
	}
	// adding optional parameters end

	sess, _ := session.Get("session", c)
	sess.Values["message"] = message
	sess.Values["status"] = status
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusMovedPermanently, path)
}
