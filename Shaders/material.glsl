

vec3 random_in_unit_sphere() {
    while (true){
        seed++;
        vec3 p = random3Neg();
        if (length_squared(p) >= 1) continue;

        return normalize(p);
    }    
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
float make_grid(vec3 pos){
    float res=2.0;
    float x1 = mod(pos.x,1.0);
    float xMask=step(x1,.5);
    
    float y1 = mod(pos.z,1.0);
    float yMask=step(y1,.5);
 
    return (xMask==0)^^(yMask==0)?1.0:.5;
}


vec3 refract(in vec3 uv, in vec3 n, float etai_over_etat) {
    float cos_theta = min(dot(-uv, n), 1.0);
    vec3 r_out_perp =  etai_over_etat * (uv + cos_theta*n);
    vec3 r_out_parallel = -sqrt(abs(1.0 - length_squared(r_out_perp))) * n;
    return r_out_perp + r_out_parallel;
}
float reflectance(float cosine, float ref_idx){
    // Use Schlick's approximation for reflectance.
    float r0 = (1-ref_idx) / (1+ref_idx);
    r0 = r0*r0;
    return r0 + (1-r0)*pow((1 - cosine),5);
}
struct Material{
    vec3 Albedo;
    float Fuzziness;
    int type; //0 is lambertian, 1 is metal
};

//Returns a new ray 
bool Scatter(in Material m,in Ray ray_in, inout hit_record hr, out vec3 attenuation, inout Ray ray_out){
    vec3 scatter_direction;
    bool scattered = false;
    if (m.type==0){ //lambertian diffuse
        scatter_direction = hr.normal + random_in_hemisphere(hr.normal);
        attenuation = m.Albedo;

        scattered=true;
    } else if (m.type==1) { //Metal
        vec3 reflected = reflect(normalize(ray_in.direction), hr.normal);

        scatter_direction = reflected + m.Fuzziness*random_in_unit_sphere(); 
        attenuation = m.Albedo;

        scattered=(dot(scatter_direction, hr.normal) > 0);
    } else{//Glass

        float refraction_ratio = hr.front_face ? (1.0/m.Fuzziness) : m.Fuzziness;

        attenuation=vec3(1);
        vec3 unit_dir = normalize(ray_in.direction);
        float cos_theta = min(dot(-unit_dir, hr.normal), 1.0);
        
        float sin_theta = sqrt(1.0 - cos_theta*cos_theta);
        bool cannot_refract = refraction_ratio * sin_theta > 1.0;
        vec3 direction;
        if (cannot_refract|| reflectance(cos_theta, refraction_ratio) > random()){
            scatter_direction = reflect(unit_dir, hr.normal);
        }else{
            scatter_direction = refract(unit_dir, hr.normal, refraction_ratio);
        }
        
        scattered=true;
    }

    // Catch degenerate scatter direction
    if (near_zero(scatter_direction)){
        scatter_direction = hr.normal;
    }

    ray_out = Ray(hr.p, scatter_direction);
    if (hr.mat_index==0){
        attenuation*=make_grid(hr.p);
    }
    return scattered;
}


