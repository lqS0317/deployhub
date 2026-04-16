package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	"deployhub/internal/service/approval"

	"github.com/gin-gonic/gin"
)

// ApprovalHandler 审批处理器
type ApprovalHandler struct {
	approvalSvc *approval.ApprovalService
}

// NewApprovalHandler 创建审批处理器
func NewApprovalHandler(approvalSvc *approval.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{approvalSvc: approvalSvc}
}

// List 查询审批列表
func (h *ApprovalHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)

	status := c.DefaultQuery("status", "pending")

	var approverID *uint
	if aidStr := c.Query("approver_id"); aidStr != "" {
		id, err := strconv.ParseUint(aidStr, 10, 32)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的审批人 ID")
			return
		}
		aid := uint(id)
		approverID = &aid
	}

	items, total, err := h.approvalSvc.List(page, pageSize, status, approverID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询审批列表失败")
		return
	}

	pkg.Paginated(c, items, total, page, pageSize)
}

// Get 获取审批详情
func (h *ApprovalHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的审批 ID")
		return
	}

	result, err := h.approvalSvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "审批记录不存在")
		return
	}

	c.JSON(http.StatusOK, result)
}

type approveRejectRequest struct {
	Comment string `json:"comment"`
}

// Approve 通过审批
func (h *ApprovalHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的审批 ID")
		return
	}

	var req approveRejectRequest
	_ = c.ShouldBindJSON(&req)

	approverID := middleware.GetUserID(c)

	if err := h.approvalSvc.Approve(uint(id), approverID, req.Comment); err != nil {
		switch err {
		case approval.ErrApprovalNotFound:
			pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, err.Error())
		case approval.ErrApprovalDecided:
			pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		case approval.ErrNotAuthorized:
			pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		default:
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "审批操作失败")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审批已通过"})
}

// Reject 驳回审批
func (h *ApprovalHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的审批 ID")
		return
	}

	var req approveRejectRequest
	_ = c.ShouldBindJSON(&req)

	approverID := middleware.GetUserID(c)

	if err := h.approvalSvc.Reject(uint(id), approverID, req.Comment); err != nil {
		switch err {
		case approval.ErrApprovalNotFound:
			pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, err.Error())
		case approval.ErrApprovalDecided:
			pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		case approval.ErrNotAuthorized:
			pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		default:
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "驳回操作失败")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审批已驳回"})
}

// RegisterApprovalRoutes 注册审批相关路由
func RegisterApprovalRoutes(r *gin.RouterGroup, h *ApprovalHandler) {
	approvals := r.Group("/approvals")
	{
		approvals.GET("", h.List)
		approvals.GET("/:id", h.Get)
		approvals.POST("/:id/approve", h.Approve)
		approvals.POST("/:id/reject", h.Reject)
	}
}
