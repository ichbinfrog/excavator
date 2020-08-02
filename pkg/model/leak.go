package model

import "time"

type Leak interface {
	templateTitle() string
	templateSubtitle() string
	templateAffected() string
}

type GitLeak struct {
	Commit   string    `yaml:"commit"`
	File     string    `yaml:"file"`
	Line     int       `yaml:"line"`
	Affected int       `yaml:"affected"`
	Snippet  []string  `yaml:"snippet"`
	Threat   float32   `yaml:"threat,omitempty"`
	Author   string    `yaml:"author,omitempty"`
	When     time.Time `yaml:"commit_date"`
	Rule     *Rule     `yaml:"-"`
	Repo     *Repo     `yaml:"-"`
}

func (g *GitLeak) templateTitle() string {
	return `
<h5 class="card-title">
	{{ .File }}
	<br>
	{{ .Commit }}
</h5>`
}

func (g *GitLeak) templateSubtitle() string {
	return `
<p class="card-text">Author: {{ .Author }}    |   At: {{ .When | date "2006-01-02 15:04:05"}}</p>`
}

func (g *GitLeak) templateAffected() string {
	return `
<div class="blob-container table-responsive">
	<table class="blob table-hover table-borderless">
		<tbody>
		{{- $start := . }}
		{{- range $idx, $line := .Snippet }}
		<tr>
			<td class="blob-num">{{ add $idx $start.Line }}</td>
			{{- if eq $idx $start.Affected }}
			<td class="blob-code text-warning">
			{{ $line }}
			</td>
			{{- else }}
			<td class="blob-code">
			{{ $line }}
			</td>
			{{- end }}
		</tr>
		{{- end }}
		</tbody>
	</table>
</div>
`
}

type FileLeak struct {
	File     string  `yaml:"file"`
	Line     int     `yaml:"line"`
	Affected string  `yaml:"affected"`
	Threat   float32 `yaml:"threat,omitempty"`
	Rule     *Rule   `yaml:"-"`
}

func (f *FileLeak) templateTitle() string {
	return `
<h5 class="card-title">
	{{ .File }}
</h5>
`
}

func (f *FileLeak) templateSubtitle() string {
	return ``
}

func (f *FileLeak) templateAffected() string {
	return `
<div class="blob-container table-responsive">
	<table class="blob table-hover table-borderless">
		<tbody>
		<tr>
			<td class="blob-num">{{ .Line }}</td>
			<td class="blob-code">
			{{ .Affected }}
			</td>
		</tr>
		</tbody>
	</table>
</div>
`
}
