#pragma once

#include "CoreMinimal.h"
#include "Components/ActorComponent.h"
#include "Live2DCubismAvatarComponent.generated.h"

// Forward declarations
class UTexture2D;
class UMaterialInstanceDynamic;

UCLASS(ClassGroup=(Custom), meta=(BlueprintSpawnableComponent))
class DEEPTREECHO_API ULive2DCubismAvatarComponent : public UActorComponent
{
    GENERATED_BODY()

public:
    ULive2DCubismAvatarComponent();

protected:
    virtual void BeginPlay() override;

public:
    virtual void TickComponent(float DeltaTime, ELevelTick TickType, FActorComponentTickFunction* ThisTickFunction) override;

    /** Load a Live2D model from a .model3.json file */
    UFUNCTION(BlueprintCallable, Category = "Live2D")
    void LoadLive2DModel(const FString& ModelPath);

    /** Update Live2D model parameters */
    UFUNCTION(BlueprintCallable, Category = "Live2D")
    void SetParameterValue(const FName& ParameterName, float Value);

    /** Get Live2D model parameter value */
    UFUNCTION(BlueprintCallable, Category = "Live2D")
    float GetParameterValue(const FName& ParameterName) const;

    /** Set emotional state (Valence, Arousal, Dominance) */
    UFUNCTION(BlueprintCallable, Category = "Live2D")
    void SetEmotionalState(float Valence, float Arousal, float Dominance);
    
    /** Set gaze target position */
    UFUNCTION(BlueprintCallable, Category = "Live2D")
    void SetGazeTarget(float X, float Y);
    
    /** Set head rotation angles */
    UFUNCTION(BlueprintCallable, Category = "Live2D")
    void SetHeadRotation(float AngleX, float AngleY, float AngleZ);
    
    /** Play expression by name */
    UFUNCTION(BlueprintCallable, Category = "Live2D")
    void PlayExpression(const FString& ExpressionName);
    
    /** Play motion by name */
    UFUNCTION(BlueprintCallable, Category = "Live2D")
    void PlayMotion(const FString& MotionName, bool bLoop = false);

protected:
    virtual void EndPlay(const EEndPlayReason::Type EndPlayReason) override;

private:
    // Live2D Cubism SDK Integration
    // Core Cubism model data (opaque pointers to SDK types)
    void* CubismModel;          // CubismModel* (opaque pointer)
    void* CubismMoc;            // CubismMoc* (opaque pointer)
    void* CubismUserModel;      // CubismUserModel* (opaque pointer)
    
    // Parameter management
    TMap<FName, int32> ParameterIndices;
    TArray<float> ParameterValues;
    
    // Expression and motion data
    TArray<void*> LoadedExpressions;  // Array of CubismExpression*
    TArray<void*> LoadedMotions;      // Array of CubismMotion*
    
    // Rendering resources
    TArray<FVector2D> VertexPositions;
    TArray<FVector2D> VertexUVs;
    TArray<int32> Indices;
    
    // Update state
    bool bModelLoaded;
    float DeltaTimeAccumulator;

    /** The Live2D model asset */
    UPROPERTY(Transient)
    UObject* Live2DModel;

    /** The texture for rendering the Live2D model */
    UPROPERTY(Transient)
    UTexture2D* RenderTarget;

    /** The dynamic material instance for the Live2D model */
    UPROPERTY(Transient)
    UMaterialInstanceDynamic* DynamicMaterial;
    
    // Helper methods
    void InitializeCubismModel();
    void UpdateCubismParameters(float DeltaTime);
    void RenderCubismModel();
    void CleanupCubismResources();
};
