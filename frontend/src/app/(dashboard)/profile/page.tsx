"use client";

import { useState, useRef } from "react";
import {
  useMe,
  useUpdateProfile,
  useChangePassword,
  useUploadAvatar,
} from "@/hooks/use-auth";
import { showToast } from "@/components/ui/toast";

export default function ProfilePage() {
  const { data: user, isLoading } = useMe();
  const updateProfile = useUpdateProfile();
  const changePassword = useChangePassword();
  const uploadAvatar = useUploadAvatar();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [nickname, setNickname] = useState("");
  const [phone, setPhone] = useState("");
  const [profileEditing, setProfileEditing] = useState(false);

  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const startEditing = () => {
    setNickname(user?.nickname || "");
    setPhone(user?.phone || "");
    setProfileEditing(true);
  };

  const handleSaveProfile = () => {
    updateProfile.mutate(
      { nickname, phone },
      {
        onSuccess: () => {
          showToast("资料更新成功", "success");
          setProfileEditing(false);
        },
      }
    );
  };

  const handleChangePassword = () => {
    if (newPassword !== confirmPassword) {
      showToast("两次密码输入不一致", "error");
      return;
    }
    if (newPassword.length < 6) {
      showToast("新密码至少 6 位", "error");
      return;
    }
    changePassword.mutate(
      { old_password: oldPassword, new_password: newPassword },
      {
        onSuccess: () => {
          showToast("密码修改成功", "success");
          setOldPassword("");
          setNewPassword("");
          setConfirmPassword("");
        },
      }
    );
  };

  const handleAvatarClick = () => {
    fileInputRef.current?.click();
  };

  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 2 * 1024 * 1024) {
      showToast("头像文件不能超过 2MB", "error");
      return;
    }

    const allowedTypes = ["image/jpeg", "image/png", "image/gif", "image/webp"];
    if (!allowedTypes.includes(file.type)) {
      showToast("仅支持 jpg/png/gif/webp 格式", "error");
      return;
    }

    uploadAvatar.mutate(file, {
      onSuccess: () => {
        showToast("头像更新成功", "success");
      },
    });
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">请先登录</div>
      </div>
    );
  }

  const avatarText = (user.nickname || user.username || "U").charAt(0).toUpperCase();

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">个人资料</h1>

      {/* 头像区域 */}
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <h2 className="text-lg font-semibold mb-4">头像</h2>
        <div className="flex items-center gap-4">
          <button
            onClick={handleAvatarClick}
            className="relative group"
            disabled={uploadAvatar.isPending}
          >
            {user.avatar ? (
              <img
                src={user.avatar}
                alt="头像"
                className="h-20 w-20 rounded-full object-cover"
              />
            ) : (
              <div className="flex h-20 w-20 items-center justify-center rounded-full bg-blue-600 text-white text-2xl font-bold">
                {avatarText}
              </div>
            )}
            <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity">
              <span className="text-white text-xs">更换</span>
            </div>
          </button>
          <div className="text-sm text-gray-500">
            <p>点击头像更换，支持 jpg/png/gif/webp</p>
            <p>最大 2MB</p>
          </div>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            className="hidden"
            onChange={handleAvatarChange}
          />
        </div>
      </div>

      {/* 基本信息 */}
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">基本信息</h2>
          {!profileEditing && (
            <button
              onClick={startEditing}
              className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              编辑
            </button>
          )}
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">
              用户名
            </label>
            <div className="text-gray-900">{user.username}</div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">
              邮箱
            </label>
            <div className="text-gray-900">{user.email}</div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">
              角色
            </label>
            <span
              className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-medium ${
                user.role === "admin"
                  ? "bg-purple-100 text-purple-700"
                  : "bg-gray-100 text-gray-700"
              }`}
            >
              {user.role === "admin" ? "管理员" : "成员"}
            </span>
          </div>

          {profileEditing ? (
            <>
              <div>
                <label className="block text-sm font-medium text-gray-600 mb-1">
                  昵称
                </label>
                <input
                  type="text"
                  value={nickname}
                  onChange={(e) => setNickname(e.target.value)}
                  maxLength={100}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  placeholder="输入昵称"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 mb-1">
                  手机号
                </label>
                <input
                  type="tel"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  maxLength={20}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  placeholder="输入手机号"
                />
              </div>
              <div className="flex gap-2 pt-2">
                <button
                  onClick={handleSaveProfile}
                  disabled={updateProfile.isPending}
                  className="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  {updateProfile.isPending ? "保存中..." : "保存"}
                </button>
                <button
                  onClick={() => setProfileEditing(false)}
                  className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-md hover:bg-gray-200"
                >
                  取消
                </button>
              </div>
            </>
          ) : (
            <>
              <div>
                <label className="block text-sm font-medium text-gray-600 mb-1">
                  昵称
                </label>
                <div className="text-gray-900">
                  {user.nickname || <span className="text-gray-400">未设置</span>}
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 mb-1">
                  手机号
                </label>
                <div className="text-gray-900">
                  {user.phone || <span className="text-gray-400">未设置</span>}
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      {/* 修改密码 */}
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <h2 className="text-lg font-semibold mb-4">修改密码</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">
              当前密码
            </label>
            <input
              type="password"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="输入当前密码"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">
              新密码
            </label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="至少 6 位"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">
              确认新密码
            </label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="再次输入新密码"
            />
          </div>
          <button
            onClick={handleChangePassword}
            disabled={
              changePassword.isPending ||
              !oldPassword ||
              !newPassword ||
              !confirmPassword
            }
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {changePassword.isPending ? "修改中..." : "修改密码"}
          </button>
        </div>
      </div>
    </div>
  );
}
