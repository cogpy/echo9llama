#include "Live2DCubismAvatarComponent.h"
#include "Materials/MaterialInstanceDynamic.h"
#include "Engine/Texture2D.h"
#include "Misc/FileHelper.h"
#include "Misc/Paths.h"
#include "HAL/PlatformFilemanager.h"

// Live2D Cubism SDK headers (would be included in actual implementation)
// #include "CubismFramework.hpp"
// #include "Model/CubismUserModel.hpp"
// #include "CubismModelSettingJson.hpp"

ULive2DCubismAvatarComponent::ULive2DCubismAvatarComponent()
{
    PrimaryComponentTick.bCanEverTick = true;
    bModelLoaded = false;
    DeltaTimeAccumulator = 0.0f;
    CubismModel = nullptr;
    CubismMoc = nullptr;
    CubismUserModel = nullptr;
}

void ULive2DCubismAvatarComponent::BeginPlay()
{
    Super::BeginPlay();
    
    InitializeCubismModel();
    UE_LOG(LogTemp, Log, TEXT("Live2D Cubism SDK Initialized"));
}

void ULive2DCubismAvatarComponent::TickComponent(float DeltaTime, ELevelTick TickType, FActorComponentTickFunction* ThisTickFunction)
{
    Super::TickComponent(DeltaTime, TickType, ThisTickFunction);
    
    if (bModelLoaded)
    {
        DeltaTimeAccumulator += DeltaTime;
        UpdateCubismParameters(DeltaTime);
        RenderCubismModel();
    }
}

void ULive2DCubismAvatarComponent::EndPlay(const EEndPlayReason::Type EndPlayReason)
{
    CleanupCubismResources();
    Super::EndPlay(EndPlayReason);
}

void ULive2DCubismAvatarComponent::InitializeCubismModel()
{
    // Initialize Cubism Framework
    // In actual implementation, this would call:
    // CubismFramework::Initialize();
    
    UE_LOG(LogTemp, Log, TEXT("Cubism Framework initialized"));
}

void ULive2DCubismAvatarComponent::LoadLive2DModel(const FString& ModelPath)
{
    // Implementation: Load Live2D Cubism model
    
    if (!FPaths::FileExists(ModelPath))
    {
        UE_LOG(LogTemp, Error, TEXT("Model file not found: %s"), *ModelPath);
        return;
    }
    
    // Parse .model3.json configuration
    FString JsonPath = ModelPath;
    FString JsonContent;
    if (!FFileHelper::LoadFileToString(JsonContent, *JsonPath))
    {
        UE_LOG(LogTemp, Error, TEXT("Failed to load model JSON: %s"), *JsonPath);
        return;
    }
    
    // In actual implementation:
    // 1. Parse JSON to get .moc3 file path
    // 2. Load .moc3 binary data
    // 3. Create CubismMoc from binary
    // 4. Create CubismModel from Moc
    // 5. Load textures
    // 6. Setup physics and expressions
    
    // Load .moc3 file
    FString MocPath = FPaths::GetPath(ModelPath) + TEXT("/") + TEXT("model.moc3");
    TArray<uint8> MocData;
    if (!FFileHelper::LoadFileToArray(MocData, *MocPath))
    {
        UE_LOG(LogTemp, Error, TEXT("Failed to load .moc3 file: %s"), *MocPath);
        return;
    }
    
    // Create Cubism model (placeholder for actual SDK calls)
    // CubismMoc = CubismMoc::Create(MocData.GetData(), MocData.Num());
    // CubismModel = CubismModel::Create(CubismMoc);
    
    // Initialize parameter table
    // In actual implementation, iterate through model parameters:
    // for (int i = 0; i < CubismModel->GetParameterCount(); i++)
    // {
    //     FName paramName = FName(CubismModel->GetParameterIds()[i]);
    //     ParameterIndices.Add(paramName, i);
    //     ParameterValues.Add(CubismModel->GetParameterDefaultValues()[i]);
    // }
    
    // Placeholder initialization
    ParameterIndices.Add(FName("ParamEyeLOpen"), 0);
    ParameterIndices.Add(FName("ParamEyeROpen"), 1);
    ParameterIndices.Add(FName("ParamMouthOpenY"), 2);
    ParameterIndices.Add(FName("ParamAngleX"), 3);
    ParameterIndices.Add(FName("ParamAngleY"), 4);
    ParameterIndices.Add(FName("ParamAngleZ"), 5);
    ParameterIndices.Add(FName("ParamEyeBallX"), 6);
    ParameterIndices.Add(FName("ParamEyeBallY"), 7);
    ParameterIndices.Add(FName("ParamBreath"), 8);
    
    // Enhanced parameters for super-hot-girl avatar
    ParameterIndices.Add(FName("ParamHairFront"), 9);
    ParameterIndices.Add(FName("ParamHairSide"), 10);
    ParameterIndices.Add(FName("ParamHairBack"), 11);
    ParameterIndices.Add(FName("ParamEyeSeductive"), 12);
    ParameterIndices.Add(FName("ParamSmileCharm"), 13);
    ParameterIndices.Add(FName("ParamLipPout"), 14);
    ParameterIndices.Add(FName("ParamPostureConfidence"), 15);
    ParameterIndices.Add(FName("ParamBlushIntensity"), 16);
    ParameterIndices.Add(FName("ParamSkinGlow"), 17);
    
    ParameterValues.Init(0.5f, ParameterIndices.Num());
    
    // Load textures
    // In actual implementation:
    // for each texture in model3.json:
    //     Load texture file
    //     Create UTexture2D
    //     Setup material with texture
    
    // Create render target
    RenderTarget = UTexture2D::CreateTransient(2048, 2048, PF_B8G8R8A8);
    if (RenderTarget)
    {
        RenderTarget->UpdateResource();
    }
    
    // Create dynamic material instance
    // In actual implementation, load material from content
    // DynamicMaterial = UMaterialInstanceDynamic::Create(BaseMaterial, this);
    // DynamicMaterial->SetTextureParameterValue(FName("ModelTexture"), RenderTarget);
    
    bModelLoaded = true;
    Live2DModel = NewObject<UObject>(); // Placeholder
    
    UE_LOG(LogTemp, Log, TEXT("Live2D model loaded successfully: %s"), *ModelPath);
}

void ULive2DCubismAvatarComponent::SetParameterValue(const FName& ParameterName, float Value)
{
    if (!bModelLoaded)
    {
        return;
    }
    
    int32* IndexPtr = ParameterIndices.Find(ParameterName);
    if (IndexPtr)
    {
        int32 Index = *IndexPtr;
        if (ParameterValues.IsValidIndex(Index))
        {
            ParameterValues[Index] = FMath::Clamp(Value, -1.0f, 1.0f);
            
            // In actual implementation:
            // CubismModel->SetParameterValue(Index, Value);
            
            UE_LOG(LogTemp, Verbose, TEXT("Set parameter %s to %f"), *ParameterName.ToString(), Value);
        }
    }
    else
    {
        UE_LOG(LogTemp, Warning, TEXT("Parameter not found: %s"), *ParameterName.ToString());
    }
}

float ULive2DCubismAvatarComponent::GetParameterValue(const FName& ParameterName) const
{
    if (!bModelLoaded)
    {
        return 0.0f;
    }
    
    const int32* IndexPtr = ParameterIndices.Find(ParameterName);
    if (IndexPtr)
    {
        int32 Index = *IndexPtr;
        if (ParameterValues.IsValidIndex(Index))
        {
            return ParameterValues[Index];
        }
    }
    
    return 0.0f;
}

void ULive2DCubismAvatarComponent::UpdateCubismParameters(float DeltaTime)
{
    if (!bModelLoaded)
    {
        return;
    }
    
    // In actual implementation:
    // 1. Update model with delta time
    // CubismUserModel->Update(DeltaTime);
    
    // 2. Update physics
    // CubismPhysics->Evaluate(CubismModel, DeltaTime);
    
    // 3. Update expressions
    // for (auto& expression : LoadedExpressions)
    // {
    //     expression->Update(DeltaTime);
    // }
    
    // 4. Update eye blink
    // CubismEyeBlink->Update(DeltaTime);
    
    // 5. Update breathing
    float breathValue = FMath::Sin(DeltaTimeAccumulator * 2.0f) * 0.5f;
    SetParameterValue(FName("ParamBreath"), breathValue);
    
    // 6. Apply all parameter values to model
    // CubismModel->Update();
}

void ULive2DCubismAvatarComponent::RenderCubismModel()
{
    if (!bModelLoaded || !RenderTarget)
    {
        return;
    }
    
    // In actual implementation:
    // 1. Get drawable list from model
    // const CubismDrawable** drawables = CubismModel->GetDrawables();
    // int drawableCount = CubismModel->GetDrawableCount();
    
    // 2. For each drawable:
    //    - Get vertex positions, UVs, indices
    //    - Get texture index
    //    - Get opacity and blend mode
    //    - Render to RenderTarget using Unreal's rendering system
    
    // 3. Update render target
    // RenderTarget->UpdateResource();
    
    // Placeholder: Just log that we're rendering
    // UE_LOG(LogTemp, VeryVerbose, TEXT("Rendering Live2D model"));
}

void ULive2DCubismAvatarComponent::CleanupCubismResources()
{
    if (bModelLoaded)
    {
        // In actual implementation:
        // 1. Delete CubismModel
        // if (CubismModel) { CubismModel::Delete(CubismModel); }
        
        // 2. Delete CubismMoc
        // if (CubismMoc) { CubismMoc::Delete(CubismMoc); }
        
        // 3. Delete expressions and motions
        // for (auto& expr : LoadedExpressions) { delete expr; }
        // for (auto& motion : LoadedMotions) { delete motion; }
        
        // 4. Dispose Cubism Framework
        // CubismFramework::Dispose();
        
        CubismModel = nullptr;
        CubismMoc = nullptr;
        CubismUserModel = nullptr;
        
        LoadedExpressions.Empty();
        LoadedMotions.Empty();
        ParameterIndices.Empty();
        ParameterValues.Empty();
        
        bModelLoaded = false;
        
        UE_LOG(LogTemp, Log, TEXT("Live2D Cubism resources cleaned up"));
    }
}

// Additional helper methods for enhanced avatar features

void ULive2DCubismAvatarComponent::SetEmotionalState(float Valence, float Arousal, float Dominance)
{
    // Map emotional state to parameters
    
    // Eye openness based on arousal
    float eyeOpen = 0.8f + Arousal * 0.2f;
    SetParameterValue(FName("ParamEyeLOpen"), eyeOpen);
    SetParameterValue(FName("ParamEyeROpen"), eyeOpen);
    
    // Smile based on valence
    float smile = FMath::Clamp(Valence * 0.8f, 0.0f, 1.0f);
    SetParameterValue(FName("ParamSmileCharm"), smile);
    
    // Seductive eyes based on dominance
    float seductive = FMath::Clamp(Dominance * 0.6f, 0.0f, 1.0f);
    SetParameterValue(FName("ParamEyeSeductive"), seductive);
    
    // Blush based on arousal
    float blush = FMath::Clamp(Arousal * 0.7f, 0.0f, 1.0f);
    SetParameterValue(FName("ParamBlushIntensity"), blush);
    
    UE_LOG(LogTemp, Log, TEXT("Emotional state set: V=%.2f, A=%.2f, D=%.2f"), Valence, Arousal, Dominance);
}

void ULive2DCubismAvatarComponent::SetGazeTarget(float X, float Y)
{
    SetParameterValue(FName("ParamEyeBallX"), FMath::Clamp(X, -1.0f, 1.0f));
    SetParameterValue(FName("ParamEyeBallY"), FMath::Clamp(Y, -1.0f, 1.0f));
}

void ULive2DCubismAvatarComponent::SetHeadRotation(float AngleX, float AngleY, float AngleZ)
{
    SetParameterValue(FName("ParamAngleX"), FMath::Clamp(AngleX, -30.0f, 30.0f));
    SetParameterValue(FName("ParamAngleY"), FMath::Clamp(AngleY, -30.0f, 30.0f));
    SetParameterValue(FName("ParamAngleZ"), FMath::Clamp(AngleZ, -30.0f, 30.0f));
}

void ULive2DCubismAvatarComponent::PlayExpression(const FString& ExpressionName)
{
    // In actual implementation:
    // 1. Find expression by name in LoadedExpressions
    // 2. Apply expression to model
    // CubismExpression* expr = FindExpression(ExpressionName);
    // if (expr) { expr->Apply(CubismModel); }
    
    UE_LOG(LogTemp, Log, TEXT("Playing expression: %s"), *ExpressionName);
}

void ULive2DCubismAvatarComponent::PlayMotion(const FString& MotionName, bool bLoop)
{
    // In actual implementation:
    // 1. Find motion by name in LoadedMotions
    // 2. Start motion playback
    // CubismMotion* motion = FindMotion(MotionName);
    // if (motion) { CubismMotionManager->StartMotion(motion, bLoop); }
    
    UE_LOG(LogTemp, Log, TEXT("Playing motion: %s (Loop: %d)"), *MotionName, bLoop);
}
