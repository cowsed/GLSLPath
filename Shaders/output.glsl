#version 330
in vec2 UV;

layout(location = 0) out vec4 frag_colour;

uniform vec2 windowDimensions;
uniform vec3 origin=vec3(0,0,0);
uniform vec3[12] spherePositions; 
uniform float[12] sphereRadii; 
uniform int[12] sphereMaterials; 

uniform vec3[10] materialColors;
uniform float[10] materialFuzziness;
uniform int[10] materialTypes;

uniform float far_distance = 10000.0;

uniform float focal_length = 1.0;

uniform float frame;
uniform float sameFrame;//Count of sameFrame

uniform int maxBounces = 2;
uniform int SamplesPerFrame=3;

uniform int render_stage=0;//0 to display texture,1 for path trace, 2 for ids

uniform sampler2D previousResult;

float seed = 1;
float random2 (vec2 uv){
    seed++;
    //seed=fract(sin(dot(uv,vec2(12.9898,78.233)))*43758.5453123);
    return fract(sin(dot(uv+seed,vec2(12.9898,78.233)))*43758.5453123);
}


//Makes a random float [0,1]
float random (){
    seed++;
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


Ray get_ray(vec3 origin, vec2 uv){
    //Image
    float aspect_ratio = windowDimensions.x/windowDimensions.y;

    //Camera
    float viewport_height = 2.0;
    float viewport_width = aspect_ratio*viewport_height;

    vec3 horizontal = vec3(viewport_width,0,0);
    vec3 vertical = vec3(0,viewport_height, 0);
    vec3 lower_left_corner = origin - horizontal/2 - vertical/2 - vec3(0, 0, focal_length);

    
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
    //Colors the ray like the sky
    float t = 0.5*(r.direction.y + 1.0);
    return (1.0-t)*vec3(1.0, 1.0, 1.0) + t*vec3(0.5, 0.7, 1.0);
    
}


vec3 random_in_unit_sphere() {
        seed++;
        vec3 p = random3Neg();
        return normalize(p);
    
}

vec3 random_unit_vector(){
    return normalize(random_in_unit_sphere());
}

vec3 random_in_hemisphere(in vec3 normal) {
    vec3 in_unit_sphere = random_in_unit_sphere();
    if (dot(in_unit_sphere, normal) > 0.0){ // In the same hemisphere as the normal
        return in_unit_sphere;
    }else{
        return -in_unit_sphere;
    }
}

bool near_zero(vec3 e) {
        // Return true if the vector is close to zero in all dimensions.
    float s = 1e-8;
    return (abs(e.x) < s) && (abs(e.y) < s) && (abs(e.z) < s);
}


struct Material{
    vec3 Albedo;
    float Fuzziness;
    int type; //0 is lambertian, 1 is metal
};

//Returns a new ray 
bool Scatter(in Material m, Ray ray_in, inout hit_record hr, out vec3 attenuation, inout Ray ray_out){
    vec3 scatter_direction;

    if (m.type==0){
        scatter_direction = hr.normal + random_unit_vector();
    } else {
        vec3 reflected = reflect(normalize(ray_in.direction), hr.normal);
        scatter_direction = reflected + m.Fuzziness*random_unit_vector();
    }

    // Catch degenerate scatter direction
    if (near_zero(scatter_direction)){
        scatter_direction = hr.normal;
    }

    ray_out = Ray(hr.p, scatter_direction);
    attenuation = m.Albedo;
    //if (hr.mat_index==0){
    //    c*=make_grid(hr.p);
    //}
    return true;
}


float make_grid(vec3 pos){
    float res=2.0;
    float x1 = mod(pos.x,1.0);
    float xMask=step(x1,.5);
    
    float y1 = mod(pos.z,1.0);
    float yMask=step(y1,.5);
 
    return (xMask==0)^^(yMask==0)?1.0:.5;//clamp(xMask+yMask,0,1)*.5+.5;
}float hit_sphere(in Ray r, in Sphere s) {
    vec3 oc = r.origin - s.center;
    float a = length_squared(r.direction);
    float half_b = dot(oc, r.direction);
    float c = length_squared(oc) - s.radius*s.radius;
    float discriminant = half_b*half_b - a*c;

    if (discriminant < 0) {
        return -1.0;
    } else {
        return (-half_b - sqrt(discriminant) ) / a;
    }
}


hit_record HitAllSpheres(in Ray r){
    hit_record hr=   hit_record(RayAt(r,far_distance), vec3(0), far_distance, true,0,0);
    float t;
    for (int i=0; i<spherePositions.length(); i++){
        
        Sphere s = Sphere(spherePositions[i], sphereRadii[i]);
        t = hit_sphere(r, s);
        if (t<hr.t && t>0.00001){
            vec3 p=RayAt(r,t);
            hr.p=p;
            hr.t=t;
            vec3 out_n = normalize(p-s.center);

            //hr.normal=out_n;
            set_face_normal(hr, r, out_n);
            hr.mat_index = sphereMaterials[i];
            hr.object_index = i;
        }


    }
    return hr;
}


vec3 ray_color(Ray initial_r, int depth){
    hit_record hr;

    Ray r = initial_r;

    float fractionalPart = 1;
    


    vec3 total_color = vec3(1);
    for (int i=depth; i>0; i--){
        hr = HitAllSpheres(r);

        //Find Ids
        if (render_stage==1){
            if (hr.t==far_distance){
                return vec3(1,0,0);//Sky's id is -1
            } else {
                return vec3(0,0,float(hr.object_index)/12.0);//intBitsToFloat
            }
        }

        if (hr.t==far_distance){
            total_color*=ColorSkyRay(r);
            break;
        }

        //Hit Object

        Ray scattered;

        vec3 c;

        Material mat = Material(materialColors[hr.mat_index],materialFuzziness[hr.mat_index],materialTypes[hr.mat_index]);

        bool reflected = Scatter(mat,r,hr,c, scattered);
        
        if (!reflected){
            break;
        }
        total_color*=c;
        r=scattered;

    }


//  else {
//         vec3 attenuation=vec3(0,0,0);
//         Ray scattered=Ray(vec3(0,0,0),vec3(0,0,0));
//         //Material mat = Material(vec3(1,0,0));//materialColors[hr.mat_index] 
//         if (true){//Scatter(mat,r,hr,attenuation,scattered)
//             return attenuation*ray_color(scattered,depth-1);
//         }
//         return vec3(0);
//     }

    //Hit nothing, return sky color
    return total_color;
}

void main() {
    vec2 uv = (UV+1)/2; //Get uv back to [0,1]x[0,1]
    if (render_stage==0){
        frag_colour = texture(previousResult, uv);
        return;
    } else if (render_stage==1){
        Ray r1 = get_ray(origin, uv);
        vec3 col;
        col=ray_color(r1,maxBounces);
        //col.xyz=vec3(1);
        //col.z/=10.0;//divide ids by total number of them for viewing pleasure
        //col.x=col.z;
        //col.y=col.z;
        //col.x=1;
        frag_colour=vec4(col, 1);
        //frag_colour=vec4(uv.x,0,1,1);
        return;
    }




    seed=random2(uv)+sameFrame;
    seed+=sin(frame);
    Ray r1;

    vec3 col=vec3(0);
    for (int i=0; i<SamplesPerFrame;i++){
        seed++;
        vec2 uv_extra = random2(uv+float(i))/windowDimensions;
        r1 = get_ray(origin, uv+uv_extra);
        col+=ray_color(r1,maxBounces);
    }

    col/=float(SamplesPerFrame);
    //Gamma correct
    col.x=sqrt(col.x);
    col.y=sqrt(col.y);    
    col.z=sqrt(col.z);
    vec3 oldCol = texture(previousResult,uv).xyz;
        
    float oldAmt = (sameFrame-1)/sameFrame;
    float newAmt = (1)/sameFrame;
    vec3 newCol = oldAmt*oldCol + newAmt*col;

    if (sameFrame==1||sameFrame==0){
        newCol=col;
    }

    frag_colour = vec4(newCol, 1);
}