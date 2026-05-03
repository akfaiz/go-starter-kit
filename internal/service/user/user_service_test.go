package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/service/user"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

func TestUserService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "User Service Suite")
}

var _ = Describe("User Service", Label("unit", "usecase"), func() {
	var (
		userRepoMock       *mocks.MockUserRepository
		passwordHasherMock *mocks.MockPasswordHasher
		svc                domain.UserService

		ctx    context.Context
		actErr error
	)
	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		userRepoMock = mocks.NewMockUserRepository(ctrl)
		passwordHasherMock = mocks.NewMockPasswordHasher(ctrl)
		svc = user.NewService(userRepoMock, passwordHasherMock)

		ctx = context.Background()

		DeferCleanup(func() {
			ctrl.Finish()
		})
	})

	Describe("FindByID", func() {
		var (
			user *domain.User
			uid  int64
		)
		BeforeEach(func() {
			uid = 1
		})
		JustBeforeEach(func() {
			user, actErr = svc.FindByID(ctx, uid)
		})
		When("the user exists", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john.doe@example.com",
				}, nil)
			})
			It("should return the user", func() {
				Expect(actErr).To(BeNil())
				Expect(user).NotTo(BeNil())
				Expect(user.ID).To(Equal(int64(1)))
				Expect(user.Name).To(Equal("John Doe"))
				Expect(user.Email).To(Equal("john.doe@example.com"))
			})
		})
		When("the user is not found", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, domain.ErrResourceNotFound)
			})
			It("should return an error", func() {
				Expect(actErr).To(Equal(domain.ErrResourceNotFound))
				Expect(user).To(BeNil())
			})
		})
	})

	Describe("UpdateProfile", func() {
		var (
			uid       int64
			inputUser *domain.User
		)
		BeforeEach(func() {
			uid = 1
			inputUser = &domain.User{
				Name:  "Jane Doe",
				Email: "jane.doe@example.com",
			}
		})
		JustBeforeEach(func() {
			actErr = svc.UpdateProfile(ctx, uid, inputUser)
		})
		When("user is not found", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(nil, domain.ErrResourceNotFound)
			})
			It("should return an error", func() {
				Expect(actErr).To(Equal(domain.ErrResourceNotFound))
			})
		})
		When("name and email change successfully", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john.doe@example.com",
				}, nil)
				var t *time.Time
				userRepoMock.EXPECT().Update(ctx, int64(1), &domain.UserUpdate{
					Name:            omit.From("Jane Doe"),
					Email:           omit.From("jane.doe@example.com"),
					EmailVerifiedAt: omitnull.FromPtr(t),
				}).Return(nil)
			})
			It("should update the profile successfully", func() {
				Expect(actErr).To(BeNil())
			})
		})
		When("name changed but email remains the same", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:    1,
					Name:  "John Doe",
					Email: inputUser.Email,
				}, nil)
				userRepoMock.EXPECT().Update(ctx, int64(1), &domain.UserUpdate{
					Name: omit.From("Jane Doe"),
				}).Return(nil)
			})
			It("should update the name only", func() {
				Expect(actErr).To(BeNil())
			})
		})
		When("neither name nor email changed", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:    1,
					Name:  "Jane Doe",
					Email: "jane.doe@example.com",
				}, nil)
			})
			It("should do nothing and return nil", func() {
				Expect(actErr).To(BeNil())
			})
		})
		When("email already exists", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john.doe@example.com",
				}, nil)
				var t *time.Time
				userRepoMock.EXPECT().Update(ctx, int64(1), &domain.UserUpdate{
					Name:            omit.From("Jane Doe"),
					Email:           omit.From("jane.doe@example.com"),
					EmailVerifiedAt: omitnull.FromPtr(t),
				}).Return(domain.ErrEmailAlreadyExists)
			})
			It("should return a domain error", func() {
				Expect(errors.Is(actErr, domain.ErrEmailAlreadyExists)).To(BeTrue())
			})
		})
		When("there is an error during update", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john.doe@example.com",
				}, nil)
				var t *time.Time
				userRepoMock.EXPECT().Update(ctx, int64(1), &domain.UserUpdate{
					Name:            omit.From("Jane Doe"),
					Email:           omit.From("jane.doe@example.com"),
					EmailVerifiedAt: omitnull.FromPtr(t),
				}).Return(errors.New("db down"))
			})
			It("bubbles the error", func() {
				Expect(actErr).To(HaveOccurred())
			})
		})
	})

	Describe("ChangePassword", func() {
		var (
			uid             int64
			currentPassword string
			newPassword     string
		)
		BeforeEach(func() {
			uid = 1
			currentPassword = "oldpassword"
			newPassword = "newpassword"
		})
		JustBeforeEach(func() {
			actErr = svc.ChangePassword(ctx, uid, currentPassword, newPassword)
		})
		When("user is not found", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(nil, domain.ErrResourceNotFound)
			})
			It("should return an error", func() {
				Expect(actErr).To(Equal(domain.ErrResourceNotFound))
			})
		})
		When("current password does not match", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().Verify(currentPassword, "hashedpassword").Return(false, nil)
			})
			It("should return a domain error", func() {
				Expect(errors.Is(actErr, domain.ErrInvalidPassword)).To(BeTrue())
			})
		})
		When("there is an error during password verification", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().
					Verify(currentPassword, "hashedpassword").
					Return(false, errors.New("bcrypt error"))
			})
			It("bubbles the error", func() {
				Expect(actErr).To(HaveOccurred())
			})
		})
		When("password is changed successfully", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().Verify(currentPassword, "hashedpassword").Return(true, nil)
				passwordHasherMock.EXPECT().Hash(newPassword).Return("newhashedpassword", nil)
				userRepoMock.EXPECT().Update(ctx, int64(1), &domain.UserUpdate{
					Password: omit.From("newhashedpassword"),
				}).Return(nil)
			})
			It("should change the password successfully", func() {
				Expect(actErr).To(BeNil())
			})
		})
		When("there is an error during password hashing", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().Verify(currentPassword, "hashedpassword").Return(true, nil)
				passwordHasherMock.EXPECT().Hash(newPassword).Return("", errors.New("bcrypt error"))
			})
			It("bubbles the error", func() {
				Expect(actErr).To(HaveOccurred())
			})
		})
		When("there is an error during update", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().Verify(currentPassword, "hashedpassword").Return(true, nil)
				passwordHasherMock.EXPECT().Hash(newPassword).Return("newhashedpassword", nil)
				userRepoMock.EXPECT().Update(ctx, int64(1), &domain.UserUpdate{
					Password: omit.From("newhashedpassword"),
				}).Return(errors.New("db down"))
			})
			It("bubbles the error", func() {
				Expect(actErr).To(HaveOccurred())
			})
		})
	})

	Describe("Delete", func() {
		var (
			uid      int64
			password string
		)
		BeforeEach(func() {
			uid = 1
			password = "userpassword"
		})
		JustBeforeEach(func() {
			actErr = svc.Delete(ctx, uid, password)
		})
		When("user is not found", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(nil, domain.ErrResourceNotFound)
			})
			It("should return an error", func() {
				Expect(actErr).To(Equal(domain.ErrResourceNotFound))
			})
		})
		When("password does not match", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().Verify(password, "hashedpassword").Return(false, nil)
			})
			It("should return a domain error", func() {
				Expect(errors.Is(actErr, domain.ErrInvalidPassword)).To(BeTrue())
			})
		})
		When("there is an error during password verification", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().Verify(password, "hashedpassword").Return(false, errors.New("bcrypt error"))
			})
			It("bubbles the error", func() {
				Expect(actErr).To(HaveOccurred())
			})
		})
		When("deletion is successful", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().Verify(password, "hashedpassword").Return(true, nil)
				userRepoMock.EXPECT().Delete(ctx, int64(1)).Return(nil)
			})
			It("should delete the user successfully", func() {
				Expect(actErr).To(BeNil())
			})
		})
		When("there is an error during deletion", func() {
			BeforeEach(func() {
				userRepoMock.EXPECT().FindByID(ctx, int64(1)).Return(&domain.User{
					ID:       1,
					Password: "hashedpassword",
				}, nil)
				passwordHasherMock.EXPECT().Verify(password, "hashedpassword").Return(true, nil)
				userRepoMock.EXPECT().Delete(ctx, int64(1)).Return(errors.New("db down"))
			})
			It("bubbles the error", func() {
				Expect(actErr).To(HaveOccurred())
			})
		})
	})
})
