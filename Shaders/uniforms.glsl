#version 330
in vec2 UV;

layout(location = 0) out vec4 frag_colour;
layout(location = 1) out vec4 frag_id;

uniform vec2 windowDimensions;

uniform vec3[12] spherePositions; 
uniform float[12] sphereRadii; 
uniform int[12] sphereMaterials; 

uniform vec3[10] materialColors;
uniform float[10] materialFuzziness;
uniform int[10] materialTypes;

uniform float far_distance = 10000.0;
uniform sampler2D env_texture;

uniform vec3 origin=vec3(0,0,0);
uniform vec3 lookat=vec3(0,0,-1);
uniform vec3 camup=vec3(0,1,0);


uniform float focal_length = 1.0;
uniform float field_of_view=90;

uniform float frame;
uniform float sameFrame;//Count of sameFrame

uniform int maxBounces = 2;
uniform int SamplesPerFrame=3;

uniform int render_stage=0;//0 to display texture,1 for path trace, 2 for ids

uniform sampler2D previousResult;

float seed = 1;


#define PI 3.1415926535897932384626433832795


