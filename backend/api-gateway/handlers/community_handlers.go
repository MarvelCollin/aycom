package handlers

import (
	"github.com/gin-gonic/gin"
)

func CreateCommunity(c *gin.Context)  {}
func UpdateCommunity(c *gin.Context)  {}
func ApproveCommunity(c *gin.Context) {}
func DeleteCommunity(c *gin.Context)  {}
func GetCommunityByID(c *gin.Context) {}
func ListCommunities(c *gin.Context)  {}

func AddMember(c *gin.Context)        {}
func RemoveMember(c *gin.Context)     {}
func ListMembers(c *gin.Context)      {}
func UpdateMemberRole(c *gin.Context) {}

func AddRule(c *gin.Context)    {}
func RemoveRule(c *gin.Context) {}
func ListRules(c *gin.Context)  {}

func RequestToJoin(c *gin.Context)      {}
func ApproveJoinRequest(c *gin.Context) {}
func RejectJoinRequest(c *gin.Context)  {}
func ListJoinRequests(c *gin.Context)   {}

func CreateChat(c *gin.Context)            {}
func AddChatParticipant(c *gin.Context)    {}
func RemoveChatParticipant(c *gin.Context) {}
func ListChats(c *gin.Context)             {}
func ListChatParticipants(c *gin.Context)  {}

func SendMessage(c *gin.Context)    {}
func DeleteMessage(c *gin.Context)  {}
func UnsendMessage(c *gin.Context)  {}
func ListMessages(c *gin.Context)   {}
func SearchMessages(c *gin.Context) {}
