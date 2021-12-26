float hit_sphere(in Ray r, in Sphere s) {
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