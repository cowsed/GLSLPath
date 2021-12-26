

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
}