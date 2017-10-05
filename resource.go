package tanklets

import (
	"os"
	"io/ioutil"
	"github.com/go-gl/gl/v3.3-core/gl"
)

// Singleton
type resourceManager struct {
	shaders map[string]*Shader
	textures map[string]*Texture2D
}

var ResourceManager = &resourceManager{
	shaders: map[string]*Shader{},
	textures: map[string]*Texture2D{},
}

func (r *resourceManager) LoadShader(vertexPath, fragmentPath, name string) (*Shader, error) {
	var vertexCode, fragmentCode string

	{
		bytes, err := ioutil.ReadFile(vertexPath)
		if err != nil {
			return nil, err
		}
		vertexCode = string(bytes)

		bytes, err = ioutil.ReadFile(fragmentPath)
		if err != nil {
			return nil, err
		}
		fragmentCode = string(bytes)
	}

	shader := NewShader(vertexCode, fragmentCode)
	r.shaders[name] = shader
	return shader, nil
}

func (r *resourceManager) Shader(name string) *Shader {
	shader, ok := r.shaders[name]
	if !ok {
		panic("Shader not found")
	}
	return shader
}

func (r *resourceManager) LoadTexture(file string, name string) (*Texture2D, error) {
	texture := NewTexture()
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	texture.Generate(f)
	r.textures[name] = texture
	return texture, nil
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
