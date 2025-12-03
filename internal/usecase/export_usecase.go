// export_usecase.go

package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"qzone-history/internal/domain/entity"
	"qzone-history/internal/domain/repository"
	"qzone-history/internal/domain/usecase"
	"time"
)

type exportUseCase struct {
	momentRepo       repository.MomentRepository
	boardMessageRepo repository.BoardMessageRepository
	friendRepo       repository.FriendRepository
}

func NewExportUseCase(
	momentRepo repository.MomentRepository,
	boardMessageRepo repository.BoardMessageRepository,
	friendRepo repository.FriendRepository,
) usecase.ExportUseCase {
	return &exportUseCase{
		momentRepo:       momentRepo,
		boardMessageRepo: boardMessageRepo,
		friendRepo:       friendRepo,
	}
}

func (u *exportUseCase) ExportUserDataToJSON(ctx context.Context, userQQ string) error {
	// 获取用户数据
	moments, _ := u.momentRepo.FindByUserQQ(ctx, userQQ, -1, 0)
	boardMessages, _ := u.boardMessageRepo.FindByUserQQ(ctx, userQQ, -1, 0)
	friends, _ := u.friendRepo.FindFriendsByUserQQ(ctx, userQQ)

	// 创建导出数据结构
	exportData := struct {
		Moments       []entity.Moment
		BoardMessages []entity.BoardMessage
		Friends       []entity.Friend
	}{
		Moments:       moments,
		BoardMessages: boardMessages,
		Friends:       friends,
	}

	// 转换为JSON
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// 写入文件
	filename := fmt.Sprintf("%s_export.json", userQQ)
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

func (u *exportUseCase) ExportUserDataToExcel(ctx context.Context, userQQ string) error {
	// TODO 实现Excel导出逻辑
	return fmt.Errorf("ExportUserDataToExcel not implemented")
}

func (u *exportUseCase) ExportUserDataToHTML(ctx context.Context, userQQ string) error {
	moments, err := u.momentRepo.FindByUserQQ(ctx, userQQ, -1, 0)
	if err != nil {
		return fmt.Errorf("failed to load moments: %w", err)
	}

	boardMessages, err := u.boardMessageRepo.FindByUserQQ(ctx, userQQ, -1, 0)
	if err != nil {
		return fmt.Errorf("failed to load board messages: %w", err)
	}

	friends, err := u.friendRepo.FindFriendsByUserQQ(ctx, userQQ)
	if err != nil {
		return fmt.Errorf("failed to load friends: %w", err)
	}

	data := struct {
		UserQQ        string
		GeneratedAt   string
		Moments       []entity.Moment
		BoardMessages []entity.BoardMessage
		Friends       []entity.Friend
	}{
		UserQQ:        userQQ,
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
		Moments:       moments,
		BoardMessages: boardMessages,
		Friends:       friends,
	}

	tmpl, err := template.New("export").Parse(userExportHTMLTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	filename := fmt.Sprintf("%s_export.html", userQQ)
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

const userExportHTMLTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>QQ空间数据导出 - {{.UserQQ}}</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Arial, sans-serif; background: #f5f5f5; margin: 0; padding: 24px; }
.container { max-width: 960px; margin: 0 auto; background: #fff; padding: 32px; box-shadow: 0 2px 10px rgba(0,0,0,0.08); border-radius: 12px; }
section { margin-top: 32px; }
.card { border: 1px solid #e5e5e5; border-radius: 10px; padding: 16px; margin-bottom: 16px; background: #fafafa; }
.meta { font-size: 0.9rem; color: #666; margin-bottom: 8px; }
.images img { max-width: 120px; margin-right: 8px; margin-bottom: 8px; border-radius: 6px; }
.comments { margin-top: 12px; padding-left: 12px; border-left: 3px solid #eee; }
.comment { font-size: 0.9rem; margin-bottom: 6px; }
.badges span { display: inline-block; font-size: 0.75rem; margin-right: 8px; padding: 2px 6px; border-radius: 4px; background: #e3f2fd; color: #1976d2; }
table { width: 100%; border-collapse: collapse; }
table th, table td { border: 1px solid #e5e5e5; padding: 8px; text-align: left; }
</style>
</head>
<body>
<div class="container">
<h1>QQ空间数据导出 - {{.UserQQ}}</h1>
<p>生成时间：{{.GeneratedAt}}</p>

<section>
<h2>说说（{{len .Moments}}）</h2>
{{if .Moments}}
	{{range .Moments}}
	<div class="card">
		<div class="meta">来自 {{.SenderQQ}} · {{if .TimeText}}{{.TimeText}}{{else}}{{.Timestamp.Format "2006-01-02 15:04:05"}}{{end}}</div>
		<p>{{.Content}}</p>
		{{if .ImageURLs}}
		<div class="images">
			{{range .ImageURLs}}<img src="{{.}}" alt="moment image">{{end}}
		</div>
		{{end}}
		<div class="badges">
			<span>点赞 {{.Likes}}</span>
			<span>浏览 {{.Views}}</span>
			{{if .IsDeleted}}<span>已删除</span>{{end}}
			{{if .IsReconstructed}}<span>已重建</span>{{end}}
		</div>
		{{if .Comments}}
		<div class="comments">
			<strong>评论</strong>
			{{range .Comments}}
			<div class="comment">{{.UserQQ}} · {{if .TimeText}}{{.TimeText}}{{else}}{{.Timestamp.Format "2006-01-02 15:04:05"}}{{end}}<br>{{.Content}}</div>
			{{end}}
		</div>
		{{end}}
	</div>
	{{end}}
{{else}}
<p>暂无说说数据。</p>
{{end}}
</section>

<section>
<h2>留言板（{{len .BoardMessages}}）</h2>
{{if .BoardMessages}}
	{{range .BoardMessages}}
	<div class="card">
		<div class="meta">{{.SenderQQ}} · {{if .TimeText}}{{.TimeText}}{{else}}{{.Timestamp.Format "2006-01-02 15:04:05"}}{{end}}</div>
		<p>{{.Content}}</p>
	</div>
	{{end}}
{{else}}
<p>暂无留言板数据。</p>
{{end}}
</section>

<section>
<h2>好友（{{len .Friends}}）</h2>
{{if .Friends}}
<table>
	<thead>
	<tr>
		<th>好友QQ</th>
		<th>昵称</th>
		<th>添加时间</th>
	</tr>
	</thead>
	<tbody>
	{{range .Friends}}
	<tr>
		<td>{{.FriendQQ}}</td>
		<td>{{.Name}}</td>
		<td>{{if .AddedTime.IsZero}}-{{else}}{{.AddedTime.Format "2006-01-02 15:04:05"}}{{end}}</td>
	</tr>
	{{end}}
	</tbody>
</table>
{{else}}
<p>暂无好友数据。</p>
{{end}}
</section>

</div>
</body>
</html>`
