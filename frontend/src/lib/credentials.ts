export type Provenance = "manual" | "github" | "imported";
export type VerificationState = "unverified" | "verified";

export type Certification = { id:string; name:string; issuer:string; issueDate:string; expiryDate:string; credentialId:string; credentialUrl:string; description:string; provenance:Provenance; verification:VerificationState; position?:number; createdAt?:string; updatedAt?:string };
export type Achievement = { id:string; title:string; description:string; date:string; sourceUrl:string; provenance:Provenance; verification:VerificationState; position?:number; createdAt?:string; updatedAt?:string };
export type CertificationDraft=Certification&{key:string};export type AchievementDraft=Achievement&{key:string};
let sequence=0;const key=(kind:string)=>`${kind}-${++sequence}`;
export const newCertificationDraft=():CertificationDraft=>({key:key("new-certification"),id:"",name:"",issuer:"",issueDate:"",expiryDate:"",credentialId:"",credentialUrl:"",description:"",provenance:"manual",verification:"unverified"});
export const newAchievementDraft=():AchievementDraft=>({key:key("new-achievement"),id:"",title:"",description:"",date:"",sourceUrl:"",provenance:"manual",verification:"unverified"});
export const certificationToDraft=(item:Certification):CertificationDraft=>({...item,key:item.id});
export const achievementToDraft=(item:Achievement):AchievementDraft=>({...item,key:item.id});
export const toCertificationInput=({key:_,position:__,createdAt:___,updatedAt:____,...item}:CertificationDraft)=>item;
export const toAchievementInput=({key:_,position:__,createdAt:___,updatedAt:____,...item}:AchievementDraft)=>item;
