package aqua

import (
	"context"
	"log"

	aqua "github.com/peurisa/aqua/pkg/api/aqua"
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

type Article struct {
	gorm.Model
	Id          int64
	Title       string
	Description string
	UserId      int64
}

type User struct {
	gorm.Model
	Id       int64
	Username string
	Password []byte
}

type aquaServiceServer struct {
	aqua.UnimplementedAquaServiceServer
	*gorm.DB
}

func NewAquaServiceServer(db *gorm.DB) aqua.AquaServiceServer {
	return &aquaServiceServer{DB: db}
}

func (s *aquaServiceServer) Hello(context.Context, *aqua.Empty) (*aqua.Message, error) {
	return &aqua.Message{
		Message: "You did it",
	}, nil
}

func (s *aquaServiceServer) CreateUser(ctx context.Context, userParam *aqua.User) (*aqua.UserResponse, error) {
	db := s.DB

	/* Check duplicate username */
	var resultCheck User
	db.Where("username = ?", userParam.Username).First(&resultCheck)

	if resultCheck.Username != "" {
		responseMessage := aqua.UserResponse{
			Status:  "failed",
			Message: "Duplicate username found",
			User:    nil,
		}

		return &responseMessage, nil
	}

	/* Hash password and create data */
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(userParam.Password), bcrypt.DefaultCost)

	user := &User{
		Username: userParam.Username,
		Password: hashedPassword,
	}

	result := db.Create(&user)

	var responseMessage aqua.UserResponse

	if result.Error != nil {
		log.Fatalf("Failed while inserting data to database: %s\n", result.Error)

		responseMessage = aqua.UserResponse{
			Status:  "failed",
			Message: result.Error.Error(),
			User:    nil,
		}
	} else {
		responseMessage = aqua.UserResponse{
			Status:  "ok",
			Message: "Data has been created",
			User: &aqua.User{
				Id:       user.Id,
				Username: user.Username,
				Password: string(user.Password),
			},
		}
	}

	return &responseMessage, nil
}

func (s *aquaServiceServer) CreateArticle(ctx context.Context, articleParam *aqua.Article) (*aqua.ArticleResponse, error) {
	db := s.DB

	article := &Article{
		Title:       articleParam.Title,
		Description: articleParam.Description,
		UserId:      12,
	}

	result := db.Create(&article)

	/* Get user data */
	var userdata *aqua.User

	db.Where("id = ?", 12).First(&userdata)

	if userdata.Username == "" {
		log.Fatalf("User id not found: %d\n", userdata.GetId())
	}

	var createArticleResponse aqua.ArticleResponse

	if result.Error != nil {
		log.Fatalf("Failed while inserting data to database: %s\n", result.Error)

		createArticleResponse = aqua.ArticleResponse{
			Status:  "failed",
			Message: result.Error.Error(),
			Article: nil,
		}
	} else {
		createArticleResponse = aqua.ArticleResponse{
			Status:  "success",
			Message: "Data has been created",
			Article: &aqua.Article{
				Id:          article.Id,
				Title:       article.Title,
				Description: article.Description,
				User:        userdata,
			},
		}
	}

	return &createArticleResponse, nil
}

func (s *aquaServiceServer) GetArticle(ctx context.Context, idParam *aqua.Article) (*aqua.ArticleResponse, error) {
	db := s.DB

	/* Get article */
	id := idParam.GetId()

	var article Article
	db.Model(Article{Id: id}).First(&article)

	/* Get user data */
	var userdata *aqua.User

	db.Where("id = ?", article.UserId).First(&userdata)

	if userdata.Username == "" {
		log.Fatalf("User not found: %d\n", userdata.GetId())
	}

	var responseMessage aqua.ArticleResponse

	if article.Title == "" {
		log.Fatalf("Article not found: %d \n", id)

		responseMessage = aqua.ArticleResponse{
			Status:  "failed",
			Message: "Article not found",
			Article: nil,
		}
	} else {
		responseMessage = aqua.ArticleResponse{
			Status:  "success",
			Message: "Article found",
			Article: &aqua.Article{
				Id:          article.Id,
				Title:       article.Title,
				Description: article.Description,
				User:        userdata,
			},
		}
	}

	return &responseMessage, nil
}

func (s *aquaServiceServer) GetArticles(ctx context.Context, empty *aqua.Empty) (*aqua.ArticlesResponse, error) {
	db := s.DB

	var articles []Article

	db.Find(&articles)

	var protoArticles []*aqua.Article

	for _, article := range articles {
		var userdata *aqua.User

		db.Where("id = ?", article.UserId).First(&userdata)

		if userdata.Username == "" {
			log.Fatalf("User not found: %d\n", userdata.GetId())
		}

		protoArticles = append(protoArticles, &aqua.Article{
			Id:          article.Id,
			Title:       article.Title,
			Description: article.Description,
			User:        userdata,
		})
	}

	articlesResponse := aqua.ArticlesResponse{
		Status:   "success",
		Message:  "Articles found",
		Articles: protoArticles,
	}

	return &articlesResponse, nil
}

func (s *aquaServiceServer) GetUser(ctx context.Context, userParam *aqua.User) (*aqua.UserResponse, error) {
	db := s.DB

	/* Get user */
	id := userParam.GetId()

	var user User
	db.Model(User{Id: id}).First(&user)

	var responseMessage aqua.UserResponse

	if user.Username == "" {
		log.Fatalf("User not found: %d \n", id)

		responseMessage = aqua.UserResponse{
			Status:  "failed",
			Message: "User not found",
			User:    nil,
		}
	} else {
		responseMessage = aqua.UserResponse{
			Status:  "success",
			Message: "User found",
			User: &aqua.User{
				Id:       user.Id,
				Username: user.Username,
				Password: string(user.Password),
			},
		}
	}

	return &responseMessage, nil
}

func (s *aquaServiceServer) GetUsers(ctx context.Context, empty *aqua.Empty) (*aqua.UsersResponse, error) {
	db := s.DB

	var users []User

	db.Find(&users)

	var protoUsers []*aqua.User

	for _, user := range users {
		protoUsers = append(protoUsers, &aqua.User{
			Id:       user.Id,
			Username: user.Username,
			Password: string(user.Password),
		})
	}

	usersResponse := aqua.UsersResponse{
		Status:  "success",
		Message: "Users found",
		User:    protoUsers,
	}

	return &usersResponse, nil
}

func (s *aquaServiceServer) UpdateArticle(ctx context.Context, articleParam *aqua.Article) (*aqua.ArticleResponse, error) {
	db := s.DB

	/* Get article */
	id := articleParam.GetId()

	var article Article
	db.First(&article, id)

	/* Get user data */
	var userdata *aqua.User
	db.Where("id = ?", article.UserId).First(&userdata)

	if userdata.Username == "" {
		log.Fatalf("User not found: %d\n", userdata.GetId())
	}

	/* Update article */
	article.Title = articleParam.Title
	article.Description = articleParam.Description

	result := db.Save(&article)

	var responseMessage aqua.ArticleResponse

	if result.Error != nil {
		log.Fatalf("Failed while updating data to database: %s\n", result.Error)

		responseMessage = aqua.ArticleResponse{
			Status:  "failed",
			Message: result.Error.Error(),
			Article: nil,
		}
	} else {
		responseMessage = aqua.ArticleResponse{
			Status:  "success",
			Message: "Data has been updated",
			Article: &aqua.Article{
				Id:          article.Id,
				Title:       article.Title,
				Description: article.Description,
				User:        userdata,
			},
		}
	}

	return &responseMessage, nil
}

func (s *aquaServiceServer) UpdateUser(ctx context.Context, userParam *aqua.User) (*aqua.UserResponse, error) {
	db := s.DB

	/* Get user */
	id := userParam.GetId()

	var user User
	db.First(&user, id)

	/* Check duplicate username */
	var resultCheck User
	db.Where("username = ?", userParam.Username).First(&resultCheck)

	if resultCheck.Username != "" {
		responseMessage := aqua.UserResponse{
			Status:  "failed",
			Message: "Duplicate username found",
			User:    nil,
		}

		return &responseMessage, nil
	}

	/* Hash password */
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(userParam.Password), bcrypt.DefaultCost)

	/* Update user */
	user.Username = userParam.Username
	user.Password = hashedPassword

	result := db.Save(&user)

	var responseMessage aqua.UserResponse

	if result.Error != nil {
		log.Fatalf("Failed while updating data to database: %s\n", result.Error)

		responseMessage = aqua.UserResponse{
			Status:  "failed",
			Message: result.Error.Error(),
			User:    nil,
		}
	} else {
		responseMessage = aqua.UserResponse{
			Status:  "success",
			Message: "Data has been updated",
			User: &aqua.User{
				Id:       user.Id,
				Username: user.Username,
				Password: string(user.Password),
			},
		}
	}

	return &responseMessage, nil
}

func (s *aquaServiceServer) DeleteArticle(ctx context.Context, articleParam *aqua.Article) (*aqua.ArticleResponse, error) {
	db := s.DB

	/* Get article */
	id := articleParam.GetId()

	var article Article
	db.First(&article, id)

	/* Get user data */
	var userdata *aqua.User
	db.Where("id = ?", article.UserId).First(&userdata)

	if userdata.Username == "" {
		log.Fatalf("User not found: %d\n", userdata.GetId())
	}

	/* Delete article */
	result := db.Delete(&article)

	var responseMessage aqua.ArticleResponse

	if result.Error != nil {
		log.Fatalf("Failed while deleting data to database: %s\n", result.Error)

		responseMessage = aqua.ArticleResponse{
			Status:  "failed",
			Message: result.Error.Error(),
			Article: nil,
		}
	} else {
		responseMessage = aqua.ArticleResponse{
			Status:  "success",
			Message: "Data has been deleted",
			Article: nil,
		}
	}

	return &responseMessage, nil
}

func (s *aquaServiceServer) DeleteUser(ctx context.Context, userParam *aqua.User) (*aqua.UserResponse, error) {
	db := s.DB

	/* Get user */
	id := userParam.GetId()

	db.Delete(&User{}, id)

	responseMessage := aqua.UserResponse{
		Status:  "success",
		Message: "Data has been deleted",
		User:    nil,
	}

	return &responseMessage, nil
}

func (s *aquaServiceServer) Login(ctx context.Context, authCredentialsParam *aqua.AuthCredentials) (*aqua.AuthResponse, error) {
	db := s.DB

	/* Get user */
	var user User
	db.Where("username = ?", authCredentialsParam.Username).First(&user)

	var responseMessage aqua.AuthResponse

	if user.Username == "" {
		// log.Fatalf("User not found: %s\n", authCredentialsParam.Username)

		responseMessage = aqua.AuthResponse{
			Status:  "failed",
			Message: "User not found",
			Token:   "",
			User:    nil,
		}
	} else {
		/* Compare password */
		err := bcrypt.CompareHashAndPassword(user.Password, []byte(authCredentialsParam.Password))

		if err != nil {
			// log.Fatalf("Wrong password: %s\n", err)

			responseMessage = aqua.AuthResponse{
				Status:  "failed",
				Message: "Wrong password",
				Token:   "",
			}
		} else {
			responseMessage = aqua.AuthResponse{
				Status:  "success",
				Message: "Login success",
				Token:   "token",
				User: &aqua.User{
					Id:       user.Id,
					Username: user.Username,
					Password: string(user.Password),
				},
			}
		}
	}

	return &responseMessage, nil
}
