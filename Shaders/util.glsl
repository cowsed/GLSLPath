
float random2 (vec2 uv){
    seed++;
    //seed=fract(sin(dot(uv,vec2(12.9898,78.233)))*43758.5453123);
    return fract(sin(dot(uv+seed,vec2(12.9898,78.233)))*43758.5453123);
}


//Makes a random float [0,1]
float random (){
    seed=fract(4012.91230*(seed+1));
    return fract(sin(seed)*43758.5453123);
}
//Makes a random float [-1,1]
float randomNeg (){
    seed++;
    return 2*fract(sin(seed)*43758.5453123)-1;
}
vec3 random3(){
    seed++;
    return vec3(random(),random(),random());
}
vec3 random3Neg(){
    seed++;
    return vec3(randomNeg(),randomNeg(),randomNeg());
}


struct Ray{
    vec3 origin;
    vec3 direction;
};
vec3 RayAt(Ray r, float t){
    return r.origin+r.direction*t;
}

float length_squared(vec3 a){
    return pow(a.x,2)+pow(a.y,2)+pow(a.z,2);
}

float degrees_to_radians(float n){
    return 3.14159*n/180.0;
}

Ray get_ray(vec3 origin, vec2 uv){
    //Image
    float aspect_ratio = windowDimensions.x/windowDimensions.y;
    float vfov=field_of_view; // vertical field-of-view in degrees
    float theta = degrees_to_radians(vfov);
    float h = tan(theta/2);

    //Camera
    vec3 vup=normalize(camup);
    float viewport_height = 2.0*h;
    float viewport_width = aspect_ratio*viewport_height;

    vec3 w = normalize(origin - lookat);
    vec3 u = normalize(cross(vup, w));
    vec3 v = cross(w, u);
    vec3 horizontal = viewport_width * u;
    vec3 vertical = viewport_height * v;
    vec3 lower_left_corner = origin - horizontal/2 - vertical/2 - w;
    
    return Ray(origin, normalize(lower_left_corner + uv.x*horizontal + uv.y*vertical - origin));
    
}

struct Sphere{
    vec3 center;
    float radius;
};


struct hit_record{
    vec3 p;
    vec3 normal;
    float t;
    bool front_face;
    int mat_index;
    int object_index;
};

void set_face_normal(inout hit_record hr, in Ray r, in vec3 outward_normal){
    hr.front_face = dot(r.direction, outward_normal) < 0;
    hr.normal = hr.front_face ? outward_normal :-outward_normal; 
}
vec3 ColorSkyRay(Ray r) {
    ////Colors the ray like the sky
    float t = 0.5*(r.direction.y + 1.0);
    return (1.0-t)*vec3(1.0, 1.0, 1.0) + t*vec3(0.5, 0.7, 1.0);
    //float u=atan(r.direction.x,r.direction.z)/(2*PI);
    //float v=1-(r.direction.y+1)/2;
    //return texture(env_texture, vec2(u,v)).xyz;
}
