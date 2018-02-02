package glpers

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/jakecoffman/tanklets/client/data"
	"bytes"
)

type resourceManager struct {
	shaders map[string]*Shader
	textures map[string]*Texture2D
}

func NewResourceManager() *resourceManager {
	return &resourceManager{
		shaders: map[string]*Shader{},
		textures: map[string]*Texture2D{},
	}
}

func (r *resourceManager) LoadShader(vertexPath, fragmentPath, name string) *Shader {
	vertexCode := string(data.MustAsset("assets/shaders/"+vertexPath))
	fragmentCode := string(data.MustAsset("assets/shaders/"+fragmentPath))

	shader := NewShader(vertexCode, fragmentCode)
	r.shaders[name] = shader
	return shader
}

func (r *resourceManager) Shader(name string) *Shader {
	shader, ok := r.shaders[name]
	if !ok {
		panic("Shader not found")
	}
	return shader
}

func (r *resourceManager) LoadTexture(file string, name string) *Texture2D {
	texture := NewTexture()
	textureBytes := data.MustAsset("assets/textures/" + file)
	texture.Generate(bytes.NewReader(textureBytes))
	r.textures[name] = texture
	return texture
}

func (r *resourceManager) Texture(name string) *Texture2D {
	t, ok := r.textures[name]
	if !ok {
		panic("Texture '" + name + "' not found")
	}
	return t
}

func (r *resourceManager) Clear() {
	for _, shader := range r.shaders {
		gl.DeleteProgram(shader.ID)
	}
	for _, texture := range r.textures {
		gl.DeleteTextures(1, &texture.ID)
	}
}
